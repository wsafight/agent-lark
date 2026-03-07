package comments

import "github.com/spf13/cobra"

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "comments", Short: "文档评论管理"}
	cmd.AddCommand(newListCommand(), newAddCommand(), newReplyCommand(), newResolveCommand())
	return cmd
}
