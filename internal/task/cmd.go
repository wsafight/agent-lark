package task

import "github.com/spf13/cobra"

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "task", Short: "任务管理"}
	cmd.AddCommand(newListCommand(), newGetCommand(), newCreateCommand(), newUpdateCommand())
	return cmd
}
