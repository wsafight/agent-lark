package im

import (
	"fmt"
	"os"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/output"
)

func newMessagesListCommand() *cobra.Command {
	var chatID string
	var limit int
	var all bool
	var before string
	var cursor string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "列举消息",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.ResolveGlobalFlags(cmd)

			if chatID == "" {
				return fmt.Errorf("MISSING_FLAG：--chat-id 是必填项")
			}

			c, err := g.NewClient()
			if err != nil {
				return err
			}

			type messageItem struct {
				MessageID  string `json:"message_id"`
				MsgType    string `json:"msg_type"`
				Content    string `json:"content"`
				CreateTime string `json:"create_time"`
				SenderID   string `json:"sender_id,omitempty"`
			}

			var items []messageItem
			var nextToken string

			pageToken := cursor

			for {
				builder := larkim.NewListMessageReqBuilder().
					ContainerIdType("chat").
					ContainerId(chatID).
					PageSize(limit)

				if pageToken != "" {
					builder = builder.PageToken(pageToken)
				}
				if before != "" {
					builder = builder.EndTime(before)
				}

				resp, err := c.Client.Im.Message.List(cmd.Context(), builder.Build(), c.RequestOptions()...)
				if err != nil {
					return fmt.Errorf("API_ERROR：%s", err.Error())
				}
				if !resp.Success() {
					return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
				}

				for _, msg := range resp.Data.Items {
					item := messageItem{}
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
					if msg.Sender != nil && msg.Sender.Id != nil {
						item.SenderID = *msg.Sender.Id
					}
					items = append(items, item)
				}

				hasMore := resp.Data.HasMore != nil && *resp.Data.HasMore
				nextPageToken := ""
				if resp.Data.PageToken != nil {
					nextPageToken = *resp.Data.PageToken
				}

				if !all || !hasMore || nextPageToken == "" {
					if hasMore && nextPageToken != "" {
						nextToken = nextPageToken
					}
					break
				}
				pageToken = nextPageToken
			}

			if g.Format == "json" {
				if g.Agent {
					return output.PrintJSON(os.Stdout, PagedResponse{Items: items, NextCursor: nextToken})
				}
				return output.PrintJSON(os.Stdout, items)
			}

			for _, item := range items {
				fmt.Printf("[%s] %s (%s): %s\n", item.CreateTime, item.SenderID, item.MsgType, item.Content)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&chatID, "chat-id", "", "群聊 ID（必填）")
	cmd.Flags().IntVar(&limit, "limit", 20, "返回数量限制")
	cmd.Flags().BoolVar(&all, "all", false, "自动翻页获取全部")
	cmd.Flags().StringVar(&before, "before", "", "获取此时间戳之前的消息")
	cmd.Flags().StringVar(&cursor, "cursor", "", "分页游标")

	return cmd
}
