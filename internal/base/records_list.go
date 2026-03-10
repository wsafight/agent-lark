package base

import (
	"fmt"
	"os"
	"strings"

	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/output"
)

func newRecordsListCommand() *cobra.Command {
	var filter string
	var selectFields string
	var limit int
	var all bool
	var cursor string

	cmd := &cobra.Command{
		Use:   "list <URL>",
		Short: "列举记录",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.ResolveGlobalFlags(cmd)

			appToken, tableID, err := parseBitableURLStrict(args[0])
			if err != nil {
				return err
			}

			c, err := g.NewClient()
			if err != nil {
				return err
			}

			var fieldNames []string
			if selectFields != "" {
				for _, f := range strings.Split(selectFields, ",") {
					f = strings.TrimSpace(f)
					if f != "" {
						fieldNames = append(fieldNames, f)
					}
				}
			}

			type recordItem struct {
				RecordID string                 `json:"record_id"`
				Fields   map[string]interface{} `json:"fields"`
			}

			var items []recordItem
			var nextToken string
			pageToken := cursor

			for {
				reqBuilder := larkbitable.NewListAppTableRecordReqBuilder().
					AppToken(appToken).
					TableId(tableID).
					PageSize(limit)

				if filter != "" {
					reqBuilder = reqBuilder.Filter(filter)
				}
				if len(fieldNames) > 0 {
					reqBuilder = reqBuilder.FieldNames(strings.Join(fieldNames, ","))
				}
				if pageToken != "" {
					reqBuilder = reqBuilder.PageToken(pageToken)
				}

				req := reqBuilder.Build()

				resp, err := c.Client.Bitable.AppTableRecord.List(cmd.Context(), req, c.RequestOptions()...)
				if err != nil {
					return fmt.Errorf("API_ERROR：%s", err.Error())
				}
				if !resp.Success() {
					return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
				}

				for _, r := range resp.Data.Items {
					item := recordItem{Fields: make(map[string]interface{})}
					if r.RecordId != nil {
						item.RecordID = *r.RecordId
					}
					for k, v := range r.Fields {
						item.Fields[k] = v
					}
					items = append(items, item)
				}

				hasMore := resp.Data.HasMore != nil && *resp.Data.HasMore
				nextPageToken := ""
				if resp.Data.PageToken != nil {
					nextPageToken = *resp.Data.PageToken
				}

				if !all || !hasMore || nextPageToken == "" {
					if hasMore && nextPageToken != "" {
						nextToken = nextPageToken
					}
					break
				}
				pageToken = nextPageToken
			}

			if g.Format == "json" {
				if g.Agent {
					return output.PrintJSON(os.Stdout, PagedResponse{Items: items, NextCursor: nextToken})
				}
				return output.PrintJSON(os.Stdout, items)
			}

			for _, item := range items {
				fmt.Printf("%s\t%v\n", item.RecordID, item.Fields)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "过滤条件（JSON 字符串）")
	cmd.Flags().StringVar(&selectFields, "select", "", "选择字段（逗号分隔）")
	cmd.Flags().IntVar(&limit, "limit", 100, "返回数量限制")
	cmd.Flags().BoolVar(&all, "all", false, "自动翻页获取全部")
	cmd.Flags().StringVar(&cursor, "cursor", "", "分页游标")
	return cmd
}
