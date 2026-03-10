package perms

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/output"
)

func newCheckCommand() *cobra.Command {
	var user string

	cmd := &cobra.Command{
		Use:   "check <URL>",
		Short: "检查特定用户权限",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.ResolveGlobalFlags(cmd)

			if user == "" {
				return fmt.Errorf("MISSING_FLAG：--user 为必填项")
			}

			token, fileType := ExtractResourceToken(args[0])

			c, err := g.NewClient()
			if err != nil {
				return err
			}

			members, err := listMembers(c, token, fileType)
			if err != nil {
				return err
			}

			// Filter target user by email or member_id
			userLower := strings.ToLower(user)
			var found *memberItem
			for i, m := range members {
				if strings.ToLower(m.MemberID) == userLower ||
					strings.ToLower(m.Email) == userLower {
					found = &members[i]
					break
				}
			}

			type checkResult struct {
				User string `json:"user"`
				Perm string `json:"perm"`
			}

			if found == nil {
				if g.Format == "json" {
					return output.PrintJSON(os.Stdout, checkResult{User: user, Perm: "no_access"})
				}
				fmt.Println("no_access")
				return nil
			}

			if g.Format == "json" {
				return output.PrintJSON(os.Stdout, checkResult{User: user, Perm: found.Perm})
			}
			fmt.Println(found.Perm)
			return nil
		},
	}

	cmd.Flags().StringVar(&user, "user", "", "邮箱或 open_id（必填）")
	return cmd
}
