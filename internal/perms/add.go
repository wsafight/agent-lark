package perms

import (
	"context"
	"fmt"
	"os"
	"strings"

	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/output"
)

func newAddCommand() *cobra.Command {
	var users []string
	var role string
	var noNotify bool
	var strict bool

	cmd := &cobra.Command{
		Use:   "add <URL>",
		Short: "添加协作者",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, _ := cmdutil.ResolveTuple(cmd)
			_ = quiet

			if len(users) == 0 {
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

			type addResult struct {
				User   string `json:"user"`
				Status string `json:"status"`
				Error  string `json:"error,omitempty"`
			}
			var results []addResult
			failedCount := 0

			for _, user := range users {
				memberType := "email"
				if strings.HasPrefix(user, "ou_") {
					memberType = "openid"
				}

				notifyLark := true
				if noNotify {
					notifyLark = false
				}

				req := larkdrive.NewCreatePermissionMemberReqBuilder().
					Token(token).
					Type(fileType).
					NeedNotification(notifyLark).
					BaseMember(larkdrive.NewBaseMemberBuilder().
						MemberType(memberType).
						MemberId(user).
						Perm(role).
						Build()).
					Build()

				resp, err := c.Client.Drive.PermissionMember.Create(
					context.Background(),
					req,
					c.RequestOptions()...,
				)
				if err != nil {
					results = append(results, addResult{User: user, Status: "failed", Error: err.Error()})
					failedCount++
					continue
				}
				if !resp.Success() {
					results = append(results, addResult{User: user, Status: "failed", Error: fmt.Sprintf("[%d] %s", resp.Code, resp.Msg)})
					failedCount++
					continue
				}
				results = append(results, addResult{User: user, Status: "succeeded"})
			}

			if format == "json" {
				if err := output.PrintJSON(os.Stdout, results); err != nil {
					return err
				}
				if strict && failedCount > 0 {
					return fmt.Errorf("PARTIAL_FAILURE：%d 个成员添加失败", failedCount)
				}
				return nil
			}

			for _, r := range results {
				if r.Status == "succeeded" {
					fmt.Printf("✓ 已添加 %s（%s）\n", r.User, role)
				} else {
					fmt.Printf("✗ 添加 %s 失败：%s\n", r.User, r.Error)
				}
			}
			if strict && failedCount > 0 {
				return fmt.Errorf("PARTIAL_FAILURE：%d 个成员添加失败", failedCount)
			}
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&users, "user", nil, "成员（可重复，邮箱或 open_id）")
	cmd.Flags().StringVar(&role, "role", "view", "权限级别：view|edit|full_access")
	cmd.Flags().BoolVar(&noNotify, "no-notify", false, "不发送通知")
	cmd.Flags().BoolVar(&strict, "strict", false, "任一用户失败时返回非零退出码")
	return cmd
}
