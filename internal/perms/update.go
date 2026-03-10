package perms

import (
	"context"
	"fmt"
	"strings"

	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/cmdutil"
)

// roleLevel returns a numeric level for a permission role (lower = fewer permissions).
func roleLevel(role string) int {
	switch role {
	case "view":
		return 1
	case "edit":
		return 2
	case "full_access":
		return 3
	default:
		return 0
	}
}

func newUpdateCommand() *cobra.Command {
	var user string
	var role string
	var yes bool

	cmd := &cobra.Command{
		Use:   "update <URL>",
		Short: "修改协作者权限",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.ResolveGlobalFlags(cmd)
			globalYes, _ := cmd.Root().PersistentFlags().GetBool("yes")

			if user == "" {
				return fmt.Errorf("MISSING_FLAG：--user 为必填项")
			}
			if role == "" {
				return fmt.Errorf("MISSING_FLAG：--role 为必填项")
			}

			token, fileType := ExtractResourceToken(args[0])

			c, err := g.NewClient()
			if err != nil {
				return err
			}

			// Check current permission for downgrade detection
			members, err := listMembers(c, token, fileType)
			if err != nil {
				return err
			}

			userLower := strings.ToLower(user)
			var current *memberItem
			for i, m := range members {
				if strings.ToLower(m.MemberID) == userLower ||
					strings.ToLower(m.Email) == userLower {
					current = &members[i]
					break
				}
			}

			isDowngrade := false
			if current != nil {
				isDowngrade = roleLevel(role) < roleLevel(current.Perm)
			}

			if isDowngrade && !yes && !globalYes && !g.Agent {
				fmt.Printf("将 %s 的权限从 %s 降级为 %s，确认？[y/N]: ", user, current.Perm, role)
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

			req := larkdrive.NewUpdatePermissionMemberReqBuilder().
				Token(token).
				Type(fileType).
				MemberId(user).
				NeedNotification(true).
				BaseMember(larkdrive.NewBaseMemberBuilder().
					MemberType(memberType).
					MemberId(user).
					Perm(role).
					Build()).
				Build()

			resp, err := c.Client.Drive.PermissionMember.Update(
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

			fmt.Printf("✓ 已将 %s 的权限更新为 %s\n", user, role)
			return nil
		},
	}

	cmd.Flags().StringVar(&user, "user", "", "成员（邮箱或 open_id）")
	cmd.Flags().StringVar(&role, "role", "", "新权限：view|edit|full_access")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "跳过确认提示")
	return cmd
}
