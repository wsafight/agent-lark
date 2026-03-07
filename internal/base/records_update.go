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

func newRecordsUpdateCommand() *cobra.Command {
	var fieldsJSON string
	var fieldPairs []string

	cmd := &cobra.Command{
		Use:   "update <URL> <record_id>",
		Short: "更新记录",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)
			_ = quiet

			if fieldsJSON == "" && len(fieldPairs) == 0 {
				return fmt.Errorf("MISSING_FLAG：--fields 或 --field 至少提供一个")
			}

			fields := make(map[string]interface{})
			if fieldsJSON != "" {
				if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
					return fmt.Errorf("INVALID_JSON：--fields 解析失败：%s", err.Error())
				}
			}
			for _, pair := range fieldPairs {
				k, v, err := parseFieldPair(pair)
				if err != nil {
					return fmt.Errorf("INVALID_FIELD：%s", err.Error())
				}
				fields[k] = v
			}

			appToken, tableID := ParseBitableURL(args[0])
			if appToken == "" {
				return fmt.Errorf("INVALID_URL：无法解析多维表格 URL")
			}
			if tableID == "" {
				return fmt.Errorf("INVALID_URL：URL 中缺少 table 参数")
			}
			recordID := args[1]

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

			record := larkbitable.NewAppTableRecordBuilder().
				Fields(fields).
				Build()

			req := larkbitable.NewUpdateAppTableRecordReqBuilder().
				AppToken(appToken).
				TableId(tableID).
				RecordId(recordID).
				AppTableRecord(record).
				Build()

			resp, err := c.Client.Bitable.AppTableRecord.Update(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			updatedID := ""
			if resp.Data.Record != nil && resp.Data.Record.RecordId != nil {
				updatedID = *resp.Data.Record.RecordId
			}

			if format == "json" {
				return output.PrintJSON(os.Stdout, map[string]string{"record_id": updatedID})
			}

			fmt.Printf("记录 %s 已更新\n", updatedID)
			return nil
		},
	}

	cmd.Flags().StringVar(&fieldsJSON, "fields", "", `字段 JSON（如 '{"姓名":"李四"}'）`)
	cmd.Flags().StringArrayVar(&fieldPairs, "field", nil, `单个字段键值对，可重复使用（如 --field "姓名=李四" --field "年龄=30"）`)
	return cmd
}
