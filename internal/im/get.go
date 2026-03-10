package im

import (
	"fmt"
	"os"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/output"
)

func newMessagesGetCommand() *cobra.Command {
	var userIDType string

	cmd := &cobra.Command{
		Use:   "get <message_id>",
		Short: "获取单条消息详情",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			messageID := args[0]
			g := cmdutil.ResolveGlobalFlags(cmd)

			c, err := g.NewClient()
			if err != nil {
				return err
			}

			builder := larkim.NewGetMessageReqBuilder().MessageId(messageID)
			if userIDType != "" {
				builder = builder.UserIdType(userIDType)
			}

			resp, err := c.Client.Im.Message.Get(cmd.Context(), builder.Build(), c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			if len(resp.Data.Items) == 0 {
				return fmt.Errorf("NOT_FOUND：未找到消息 %s", messageID)
			}

			msg := resp.Data.Items[0]

			type messageDetail struct {
				MessageID  string `json:"message_id"`
				MsgType    string `json:"msg_type"`
				Content    string `json:"content"`
				CreateTime string `json:"create_time"`
				UpdateTime string `json:"update_time,omitempty"`
				SenderID   string `json:"sender_id,omitempty"`
				ChatID     string `json:"chat_id,omitempty"`
				RootID     string `json:"root_id,omitempty"`
				ParentID   string `json:"parent_id,omitempty"`
				Deleted    bool   `json:"deleted,omitempty"`
				Updated    bool   `json:"updated,omitempty"`
			}

			item := messageDetail{}
			if msg.MessageId != nil {
				item.MessageID = *msg.MessageId
			}
			if msg.MsgType != nil {
				item.MsgType = *msg.MsgType
			}
			if msg.Body != nil && msg.Body.Content != nil {
				item.Content = *msg.Body.Content
			}
			if msg.CreateTime != nil {
				item.CreateTime = *msg.CreateTime
			}
			if msg.UpdateTime != nil {
				item.UpdateTime = *msg.UpdateTime
			}
			if msg.Sender != nil && msg.Sender.Id != nil {
				item.SenderID = *msg.Sender.Id
			}
			if msg.ChatId != nil {
				item.ChatID = *msg.ChatId
			}
			if msg.RootId != nil {
				item.RootID = *msg.RootId
			}
			if msg.ParentId != nil {
				item.ParentID = *msg.ParentId
			}
			if msg.Deleted != nil {
				item.Deleted = *msg.Deleted
			}
			if msg.Updated != nil {
				item.Updated = *msg.Updated
			}

			if g.Format == "json" {
				return output.PrintJSON(os.Stdout, item)
			}

			fmt.Printf("message_id:  %s\n", item.MessageID)
			fmt.Printf("msg_type:    %s\n", item.MsgType)
			fmt.Printf("sender_id:   %s\n", item.SenderID)
			fmt.Printf("chat_id:     %s\n", item.ChatID)
			fmt.Printf("create_time: %s\n", item.CreateTime)
			if item.UpdateTime != "" {
				fmt.Printf("update_time: %s\n", item.UpdateTime)
			}
			if item.RootID != "" {
				fmt.Printf("root_id:     %s\n", item.RootID)
			}
			if item.ParentID != "" {
				fmt.Printf("parent_id:   %s\n", item.ParentID)
			}
			fmt.Printf("content:     %s\n", item.Content)
			return nil
		},
	}

	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "用户 ID 类型（open_id / union_id / user_id）")

	return cmd
}
