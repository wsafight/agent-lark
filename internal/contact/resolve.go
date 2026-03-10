package contact

import (
	"fmt"
	"strings"

	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/cmdutil"
)

func newResolveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve <邮箱>",
		Short: "通过邮箱解析 open_id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.ResolveGlobalFlags(cmd)

			input := args[0]

			if !strings.Contains(input, "@") {
				return fmt.Errorf("UNSUPPORTED：仅支持邮箱解析，请输入包含 @ 的邮箱地址")
			}

			c, err := g.NewClient()
			if err != nil {
				return err
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
				return fmt.Errorf("NOT_FOUND：未找到邮箱 %q 对应的用户", input)
			}

			fmt.Println(openID)
			return nil
		},
	}

	return cmd
}
