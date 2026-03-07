package base

import (
	"fmt"

	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/output"
)

func newRecordsCountCommand() *cobra.Command {
	var filter string

	cmd := &cobra.Command{
		Use:   "count <URL>",
		Short: "统计记录数",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
			}
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

			reqBuilder := larkbitable.NewListAppTableRecordReqBuilder().
				AppToken(appToken).
				TableId(tableID).
				PageSize(1)

			if filter != "" {
				reqBuilder = reqBuilder.Filter(filter)
			}

			req := reqBuilder.Build()

			resp, err := c.Client.Bitable.AppTableRecord.List(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			total := 0
			if resp.Data.Total != nil {
				total = *resp.Data.Total
			}

			fmt.Println(total)
			return nil
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "过滤条件（JSON 字符串）")
	return cmd
}
