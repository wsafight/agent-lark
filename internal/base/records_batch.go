package base

import (
	"encoding/json"
	"fmt"
	"os"

	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/output"
)

func newRecordsBatchCreateCommand() *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "batch-create <URL>",
		Short: "批量创建记录",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)
			_ = quiet

			if filePath == "" {
				return fmt.Errorf("MISSING_FLAG：--file 为必填项")
			}

			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("FILE_ERROR：无法读取文件 %s：%s", filePath, err.Error())
			}

			var recordsData []map[string]interface{}
			if err := json.Unmarshal(data, &recordsData); err != nil {
				return fmt.Errorf("INVALID_JSON：文件内容解析失败：%s", err.Error())
			}

			if len(recordsData) == 0 {
				return fmt.Errorf("EMPTY_DATA：文件中没有记录")
			}

			appToken, tableID := ParseBitableURL(args[0])
			if appToken == "" {
				return fmt.Errorf("INVALID_URL：无法解析多维表格 URL")
			}
			if tableID == "" {
				return fmt.Errorf("INVALID_URL：URL 中缺少 table 参数")
			}

			c, err := client.New(client.Options{
				TokenMode: tokenMode,
				Debug:     debug,
				Profile:   profile,
				Config:    cfg,
				Domain:    domain,
			})
			if err != nil {
				return fmt.Errorf("CLIENT_ERROR：%s", err.Error())
			}

			var records []*larkbitable.AppTableRecord
			for _, rd := range recordsData {
				r := larkbitable.NewAppTableRecordBuilder().Fields(rd).Build()
				records = append(records, r)
			}

			req := larkbitable.NewBatchCreateAppTableRecordReqBuilder().
				AppToken(appToken).
				TableId(tableID).
				Body(larkbitable.NewBatchCreateAppTableRecordReqBodyBuilder().
					Records(records).
					Build()).
				Build()

			resp, err := c.Client.Bitable.AppTableRecord.BatchCreate(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			var recordIDs []string
			for _, r := range resp.Data.Records {
				if r.RecordId != nil {
					recordIDs = append(recordIDs, *r.RecordId)
				}
			}

			type batchResult struct {
				Count     int      `json:"count"`
				RecordIDs []string `json:"record_ids"`
			}

			result := batchResult{Count: len(recordIDs), RecordIDs: recordIDs}

			if format == "json" {
				return output.PrintJSON(os.Stdout, result)
			}

			fmt.Printf("已创建 %d 条记录\n", result.Count)
			for _, id := range result.RecordIDs {
				fmt.Printf("  %s\n", id)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "JSON 文件路径（包含记录数组）")
	return cmd
}
