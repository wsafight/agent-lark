package im

import "github.com/spf13/cobra"

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "im", Short: "即时消息"}

	chats := &cobra.Command{Use: "chats", Short: "群聊管理"}
	chats.AddCommand(newChatsListCommand(), newChatsSearchCommand())

	messages := &cobra.Command{Use: "messages", Short: "消息管理"}
	messages.AddCommand(newMessagesListCommand())

	react := &cobra.Command{Use: "react", Short: "表情回复"}
	react.AddCommand(newReactAddCommand(), newReactRemoveCommand())

	cmd.AddCommand(chats, messages, react, newSendCommand())
	return cmd
}
