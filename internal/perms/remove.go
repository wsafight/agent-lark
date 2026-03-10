package perms

import (
	"context"
	"fmt"
	"strings"

	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/output"
)

func newRemoveCommand() *cobra.Command {
	var user string
	var yes bool

	cmd := &cobra.Command{
		Use:   "remove <URL>",
		Short: "移除协作者",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, tokenMode, profile, cfg, domain, debug, quiet, _ := cmdutil.ResolveTuple(cmd)
			_ = quiet
			globalYes, _ := cmd.Root().PersistentFlags().GetBool("yes")

			if user == "" {
				return fmt.Errorf("MISSING_FLAG：--user 为必填项")
			}

			token, fileType := ExtractResourceToken(args[0])

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

			if !yes && !globalYes && !output.GlobalAgent {
				fmt.Printf("确认移除 %s 的访问权限？[y/N]: ", user)
				var input string
				fmt.Scan(&input)
				if input != "y" {
					fmt.Println("已取消")
					return nil
				}
			}

			memberType := "email"
			if strings.HasPrefix(user, "ou_") {
				memberType = "openid"
			}

			req := larkdrive.NewDeletePermissionMemberReqBuilder().
				Token(token).
				Type(fileType).
				MemberId(user).
				MemberType(memberType).
				Build()

			resp, err := c.Client.Drive.PermissionMember.Delete(
				context.Background(),
				req,
				c.RequestOptions()...,
			)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			fmt.Printf("✓ 已移除 %s 的访问权限\n", user)
			return nil
		},
	}

	cmd.Flags().StringVar(&user, "user", "", "成员（邮箱或 open_id）")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "跳过确认提示")
	return cmd
}
