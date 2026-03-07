package im

import (
	"fmt"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/output"
)

func newReactAddCommand() *cobra.Command {
	var messageID string
	var emoji string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "添加表情回复",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
			}

			if messageID == "" {
				return fmt.Errorf("MISSING_FLAG：--message-id 是必填项")
			}
			if emoji == "" {
				return fmt.Errorf("MISSING_FLAG：--emoji 是必填项")
			}

			c, err := client.New(client.Options{
				TokenMode: tokenMode,
				Debug:     debug,
				Profile:   profile,
				Config:    cfg,
				Domain:    domain,
			})
			if err != nil {
				return fmt.Errorf("CLIENT_ERROR：%s", err.Error())
			}

			req := larkim.NewCreateMessageReactionReqBuilder().
				MessageId(messageID).
				Body(larkim.NewCreateMessageReactionReqBodyBuilder().
					ReactionType(larkim.NewEmojiBuilder().EmojiType(emoji).Build()).
					Build()).
				Build()

			resp, err := c.Client.Im.MessageReaction.Create(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			reactionID := ""
			if resp.Data.ReactionId != nil {
				reactionID = *resp.Data.ReactionId
			}

			output.PrintSuccess(quiet, fmt.Sprintf("表情已添加，reaction_id: %s", reactionID))

			if output.GlobalAgent {
				return output.PrintJSON(cmd.OutOrStdout(), map[string]string{
					"reaction_id": reactionID,
					"message_id":  messageID,
					"emoji":       emoji,
				})
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&messageID, "message-id", "", "消息 ID（必填）")
	cmd.Flags().StringVar(&emoji, "emoji", "", "表情类型（必填）")

	return cmd
}

func newReactRemoveCommand() *cobra.Command {
	var messageID string
	var reactionID string

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "删除表情回复",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
			}

			if messageID == "" {
				return fmt.Errorf("MISSING_FLAG：--message-id 是必填项")
			}
			if reactionID == "" {
				return fmt.Errorf("MISSING_FLAG：--reaction-id 是必填项")
			}

			c, err := client.New(client.Options{
				TokenMode: tokenMode,
				Debug:     debug,
				Profile:   profile,
				Config:    cfg,
				Domain:    domain,
			})
			if err != nil {
				return fmt.Errorf("CLIENT_ERROR：%s", err.Error())
			}

			req := larkim.NewDeleteMessageReactionReqBuilder().
				MessageId(messageID).
				ReactionId(reactionID).
				Build()

			resp, err := c.Client.Im.MessageReaction.Delete(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			output.PrintSuccess(quiet, fmt.Sprintf("表情已删除，reaction_id: %s", reactionID))

			if output.GlobalAgent {
				return output.PrintJSON(cmd.OutOrStdout(), map[string]string{
					"reaction_id": reactionID,
					"message_id":  messageID,
				})
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&messageID, "message-id", "", "消息 ID（必填）")
	cmd.Flags().StringVar(&reactionID, "reaction-id", "", "Reaction ID（必填）")

	return cmd
}
