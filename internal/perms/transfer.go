package perms

import (
	"context"
	"fmt"
	"strings"

	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/cmdutil"
)

func newTransferCommand() *cobra.Command {
	var to string
	var yes bool

	cmd := &cobra.Command{
		Use:   "transfer <URL>",
		Short: "转让文档所有权",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.ResolveGlobalFlags(cmd)
			globalYes, _ := cmd.Root().PersistentFlags().GetBool("yes")

			if to == "" {
				return fmt.Errorf("MISSING_FLAG：--to 为必填项")
			}

			token, fileType := ExtractResourceToken(args[0])

			c, err := g.NewClient()
			if err != nil {
				return err
			}

			if !yes && !globalYes && !g.Agent {
				fmt.Printf("⚠ 转让所有权给 %s 后不可撤销，确认？[y/N]: ", to)
				var input string
				fmt.Scan(&input)
				if input != "y" {
					fmt.Println("已取消")
					return nil
				}
			}

			memberType := "email"
			if strings.HasPrefix(to, "ou_") {
				memberType = "openid"
			}

			req := larkdrive.NewTransferOwnerPermissionMemberReqBuilder().
				Type(fileType).
				Token(token).
				RemoveOldOwner(false).
				Owner(larkdrive.NewOwnerBuilder().
					MemberType(memberType).
					MemberId(to).
					Build()).
				Build()

			resp, err := c.Client.Drive.PermissionMember.TransferOwner(
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

			fmt.Printf("✓ 已将所有权转让给 %s\n", to)
			return nil
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "新所有者（邮箱或 open_id）")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "跳过确认提示")
	return cmd
}
