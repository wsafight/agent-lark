package docs

import "github.com/spf13/cobra"

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "docs", Short: "文档操作"}
	cmd.AddCommand(
		newListCommand(),
		newGetCommand(),
		newOutlineCommand(),
		newSearchCommand(),
		newCreateCommand(),
		newUpdateCommand(),
		newTableCommand(),
	)
	return cmd
}
