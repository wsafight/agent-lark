package im

import (
	"encoding/json"
	"fmt"
	"os"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/output"
)

func newSendCommand() *cobra.Command {
	var chatID string
	var userID string
	var text string
	var cardFile string
	var cardJSON string

	cmd := &cobra.Command{
		Use:   "send",
		Short: "发送消息",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.ResolveGlobalFlags(cmd)

			c, err := g.NewClient()
			if err != nil {
				return err
			}

			// Determine receive ID and type
			receiveIDType := "chat_id"
			receiveID := chatID

			if userID != "" {
				// When sending to a user, use open_id receive type directly.
				// The Feishu API supports sending to a user by open_id directly,
				// but to create a p2p chat first, we create the chat.
				createResp, err := c.Client.Im.Chat.Create(
					cmd.Context(),
					larkim.NewCreateChatReqBuilder().
						UserIdType("open_id").
						Body(
							larkim.NewCreateChatReqBodyBuilder().
								UserIdList([]string{userID}).
								Build(),
						).
						Build(),
					c.RequestOptions()...,
				)
				if err != nil {
					return fmt.Errorf("API_ERROR：创建单聊失败：%s", err.Error())
				}
				if !createResp.Success() {
					return fmt.Errorf("API_ERROR：创建单聊失败：[%d] %s", createResp.Code, createResp.Msg)
				}
				if createResp.Data.ChatId != nil {
					receiveID = *createResp.Data.ChatId
				}
				receiveIDType = "chat_id"
			}

			if receiveID == "" {
				return fmt.Errorf("MISSING_FLAG：需要 --chat-id 或 --user-id 之一")
			}

			// Determine message type and content
			msgType := "text"
			msgContent := ""

			if cardFile != "" {
				data, err := os.ReadFile(cardFile)
				if err != nil {
					return fmt.Errorf("FILE_ERROR：%s", err.Error())
				}
				msgType = "interactive"
				msgContent = string(data)
			} else if cardJSON != "" {
				msgType = "interactive"
				msgContent = cardJSON
			} else if text != "" {
				msgType = "text"
				b, _ := json.Marshal(map[string]string{"text": text})
				msgContent = string(b)
			} else {
				return fmt.Errorf("MISSING_CONTENT：需要 --text、--card-file 或 --card 之一")
			}

			req := larkim.NewCreateMessageReqBuilder().
				ReceiveIdType(receiveIDType).
				Body(
					larkim.NewCreateMessageReqBodyBuilder().
						ReceiveId(receiveID).
						MsgType(msgType).
						Content(msgContent).
						Build(),
				).
				Build()

			resp, err := c.Client.Im.Message.Create(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			messageID := ""
			if resp.Data.MessageId != nil {
				messageID = *resp.Data.MessageId
			}

			output.PrintSuccess(g.Quiet, fmt.Sprintf("消息已发送，message_id: %s", messageID))

			if g.Agent {
				return output.PrintJSON(cmd.OutOrStdout(), map[string]string{
					"message_id": messageID,
					"chat_id":    receiveID,
				})
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&chatID, "chat-id", "", "群聊 ID")
	cmd.Flags().StringVar(&userID, "user-id", "", "用户 open_id（自动创建单聊）")
	cmd.Flags().StringVar(&text, "text", "", "文本消息内容")
	cmd.Flags().StringVar(&cardFile, "card-file", "", "卡片消息 JSON 文件路径")
	cmd.Flags().StringVar(&cardJSON, "card", "", "卡片消息 JSON 字符串")

	return cmd
}

