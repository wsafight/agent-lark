package base

import (
	"fmt"
	"os"

	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/output"
)

func newRecordsGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <URL> <record_id>",
		Short: "获取单条记录",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)
			_ = quiet

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

			req := larkbitable.NewGetAppTableRecordReqBuilder().
				AppToken(appToken).
				TableId(tableID).
				RecordId(recordID).
				Build()

			resp, err := c.Client.Bitable.AppTableRecord.Get(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			type recordItem struct {
				RecordID string                 `json:"record_id"`
				Fields   map[string]interface{} `json:"fields"`
			}

			item := recordItem{Fields: make(map[string]interface{})}
			if resp.Data.Record != nil {
				if resp.Data.Record.RecordId != nil {
					item.RecordID = *resp.Data.Record.RecordId
				}
				for k, v := range resp.Data.Record.Fields {
					item.Fields[k] = v
				}
			}

			if format == "json" {
				return output.PrintJSON(os.Stdout, item)
			}

			fmt.Printf("record_id: %s\n", item.RecordID)
			for k, v := range item.Fields {
				fmt.Printf("  %s: %v\n", k, v)
			}
			return nil
		},
	}

	return cmd
}
