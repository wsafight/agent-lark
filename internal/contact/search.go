package contact

import (
	"fmt"
	"os"
	"strings"

	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/output"
)

type userItem struct {
	OpenID     string `json:"open_id"`
	Name       string `json:"name"`
	Email      string `json:"email,omitempty"`
	Department string `json:"department,omitempty"`
}

func newSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <邮箱>",
		Short: "通过邮箱查询用户",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)
			_ = quiet

			query := args[0]

			if !strings.Contains(query, "@") {
				return fmt.Errorf("UNSUPPORTED：仅支持邮箱搜索，请输入包含 @ 的邮箱地址")
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

			req := larkcontact.NewBatchGetIdUserReqBuilder().
				UserIdType("open_id").
				Body(larkcontact.NewBatchGetIdUserReqBodyBuilder().
					Emails([]string{query}).
					Build()).
				Build()

			resp, err := c.Client.Contact.User.BatchGetId(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			var items []userItem
			for _, u := range resp.Data.UserList {
				item := userItem{}
				if u.UserId != nil {
					item.OpenID = *u.UserId
				}
				if u.Email != nil {
					item.Email = *u.Email
				}
				if item.OpenID != "" || item.Email != "" {
					items = append(items, item)
				}
			}

			if format == "json" {
				return output.PrintJSON(os.Stdout, items)
			}

			for _, item := range items {
				fmt.Printf("%s\t%s\n", item.OpenID, item.Email)
			}
			return nil
		},
	}

	return cmd
}
