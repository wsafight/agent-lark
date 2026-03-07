package contact

import "github.com/spf13/cobra"

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "contact", Short: "通讯录"}
	cmd.AddCommand(newSearchCommand(), newListCommand(), newResolveCommand())
	return cmd
}
