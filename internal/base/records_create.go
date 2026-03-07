package base

import (
	"encoding/json"
	"fmt"
	"os"

	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
	"github.com/spf13/cobra"
	"github.com/wangshian/agent-lark/internal/client"
	"github.com/wangshian/agent-lark/internal/output"
)

func newRecordsCreateCommand() *cobra.Command {
	var fieldsJSON string

	cmd := &cobra.Command{
		Use:   "create <URL>",
		Short: "创建记录",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)
			_ = quiet

			if fieldsJSON == "" {
				return fmt.Errorf("MISSING_FLAG：--fields 为必填项")
			}

			var fields map[string]interface{}
			if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
				return fmt.Errorf("INVALID_JSON：--fields 解析失败：%s", err.Error())
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

			record := larkbitable.NewAppTableRecordBuilder().
				Fields(fields).
				Build()

			req := larkbitable.NewCreateAppTableRecordReqBuilder().
				AppToken(appToken).
				TableId(tableID).
				AppTableRecord(record).
				Build()

			resp, err := c.Client.Bitable.AppTableRecord.Create(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			recordID := ""
			if resp.Data.Record != nil && resp.Data.Record.RecordId != nil {
				recordID = *resp.Data.Record.RecordId
			}

			if format == "json" {
				return output.PrintJSON(os.Stdout, map[string]string{"record_id": recordID})
			}

			fmt.Println(recordID)
			return nil
		},
	}

	cmd.Flags().StringVar(&fieldsJSON, "fields", "", `字段 JSON（如 '{"姓名":"张三"}'）`)
	return cmd
}
