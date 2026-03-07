package contact

import (
	"fmt"
	"os"

	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	"github.com/spf13/cobra"
	"github.com/wangshian/agent-lark/internal/client"
	"github.com/wangshian/agent-lark/internal/output"
)

func newListCommand() *cobra.Command {
	var dept string
	var limit int
	var all bool
	var cursor string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "列举用户",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)
			_ = quiet

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

			deptID := "0" // root department
			if dept != "" {
				// Search for department by name first
				deptSearchReq := larkcontact.NewSearchDepartmentReqBuilder().
					Body(larkcontact.NewSearchDepartmentReqBodyBuilder().
						Query(dept).
						Build()).
					Build()

				deptSearchResp, err := c.Client.Contact.Department.Search(cmd.Context(), deptSearchReq, c.RequestOptions()...)
				if err != nil {
					return fmt.Errorf("API_ERROR：查询部门失败：%s", err.Error())
				}
				if !deptSearchResp.Success() {
					return fmt.Errorf("API_ERROR：[%d] %s", deptSearchResp.Code, deptSearchResp.Msg)
				}
				if len(deptSearchResp.Data.Items) == 0 {
					return fmt.Errorf("NOT_FOUND：未找到部门 %q", dept)
				}
				if deptSearchResp.Data.Items[0].DepartmentId != nil {
					deptID = *deptSearchResp.Data.Items[0].DepartmentId
				}
			}

			var items []userItem
			var nextToken string
			pageToken := cursor

			for {
				reqBuilder := larkcontact.NewListUserReqBuilder().
					DepartmentId(deptID).
					PageSize(limit)

				if pageToken != "" {
					reqBuilder = reqBuilder.PageToken(pageToken)
				}

				req := reqBuilder.Build()

				resp, err := c.Client.Contact.User.List(cmd.Context(), req, c.RequestOptions()...)
				if err != nil {
					return fmt.Errorf("API_ERROR：%s", err.Error())
				}
				if !resp.Success() {
					return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
				}

				for _, u := range resp.Data.Items {
					item := userItem{}
					if u.OpenId != nil {
						item.OpenID = *u.OpenId
					}
					if u.Name != nil {
						item.Name = *u.Name
					}
					if u.Email != nil {
						item.Email = *u.Email
					}
					if len(u.DepartmentIds) > 0 {
						item.Department = u.DepartmentIds[0]
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

			if format == "json" {
				if agent {
					return output.PrintJSON(os.Stdout, PagedResponse{Items: items, NextCursor: nextToken})
				}
				return output.PrintJSON(os.Stdout, items)
			}

			for _, item := range items {
				fmt.Printf("%s\t%s\t%s\t%s\n", item.Name, item.OpenID, item.Email, item.Department)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dept, "dept", "", "部门名称")
	cmd.Flags().IntVar(&limit, "limit", 50, "返回数量限制")
	cmd.Flags().BoolVar(&all, "all", false, "自动翻页获取全部")
	cmd.Flags().StringVar(&cursor, "cursor", "", "分页游标")
	return cmd
}
