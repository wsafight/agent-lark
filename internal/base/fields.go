package base

import (
	"fmt"
	"os"
	"strings"

	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/output"
)

func newFieldsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fields <URL>",
		Short: "查看表格字段结构",
		Args:  cobra.ExactArgs(1),
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

			req := larkbitable.NewListAppTableFieldReqBuilder().
				AppToken(appToken).
				TableId(tableID).
				Build()

			resp, err := c.Client.Bitable.AppTableField.List(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			type fieldItem struct {
				FieldID   string   `json:"field_id"`
				FieldName string   `json:"field_name"`
				FieldType int      `json:"field_type"`
				Required  bool     `json:"required"`
				Options   []string `json:"options,omitempty"`
			}

			var items []fieldItem
			for _, f := range resp.Data.Items {
				item := fieldItem{}
				if f.FieldId != nil {
					item.FieldID = *f.FieldId
				}
				if f.FieldName != nil {
					item.FieldName = *f.FieldName
				}
				if f.Type != nil {
					item.FieldType = *f.Type
				}
				if f.Property != nil && f.Property.Options != nil {
					for _, opt := range f.Property.Options {
						if opt.Name != nil {
							item.Options = append(item.Options, *opt.Name)
						}
					}
				}
				items = append(items, item)
			}

			if format == "json" {
				return output.PrintJSON(os.Stdout, items)
			}

			for _, item := range items {
				required := ""
				if item.Required {
					required = "必填"
				}
				optStr := ""
				if len(item.Options) > 0 {
					optStr = fmt.Sprintf("  [%s]", strings.Join(item.Options, ", "))
				}
				fmt.Printf("%s\t类型:%d\t%s%s\n", item.FieldName, item.FieldType, required, optStr)
			}
			return nil
		},
	}

	return cmd
}
