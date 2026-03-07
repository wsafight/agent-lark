package wiki

import "github.com/spf13/cobra"

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "wiki", Short: "知识库操作"}
	cmd.AddCommand(newListCommand(), newGetCommand(), newCreateCommand())
	return cmd
}
