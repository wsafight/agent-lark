package base

import "github.com/spf13/cobra"

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "base", Short: "多维表格操作"}
	records := &cobra.Command{Use: "records", Short: "记录管理"}
	records.AddCommand(
		newRecordsListCommand(),
		newRecordsGetCommand(),
		newRecordsCreateCommand(),
		newRecordsUpdateCommand(),
		newRecordsBatchCreateCommand(),
		newRecordsCountCommand(),
	)
	cmd.AddCommand(records, newFieldsCommand())
	return cmd
}
