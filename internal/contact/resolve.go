package contact

import (
	"fmt"
	"strings"

	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	"github.com/spf13/cobra"
	"github.com/wangshian/agent-lark/internal/client"
	"github.com/wangshian/agent-lark/internal/output"
)

func newResolveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve <邮箱>",
		Short: "通过邮箱解析 open_id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
			}
			_ = quiet

			input := args[0]

			if !strings.Contains(input, "@") {
				return fmt.Errorf("UNSUPPORTED：仅支持邮箱解析，请输入包含 @ 的邮箱地址")
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
					Emails([]string{input}).
					Build()).
				Build()

			resp, err := c.Client.Contact.User.BatchGetId(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			var openID string
			for _, u := range resp.Data.UserList {
				if u.UserId != nil && *u.UserId != "" {
					openID = *u.UserId
					break
				}
			}

			if openID == "" {
				output.PrintErrorCode("NOT_FOUND", fmt.Sprintf("未找到邮箱 %q 对应的用户", input), "")
				return fmt.Errorf("NOT_FOUND：未找到用户")
			}

			fmt.Println(openID)
			return nil
		},
	}

	return cmd
}
