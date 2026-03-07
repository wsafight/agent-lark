package perms

import "github.com/spf13/cobra"

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "perms", Short: "文档权限管理"}
	cmd.AddCommand(
		newListCommand(),
		newCheckCommand(),
		newAddCommand(),
		newUpdateCommand(),
		newRemoveCommand(),
		newTransferCommand(),
		newPublicCommand(),
	)
	return cmd
}
