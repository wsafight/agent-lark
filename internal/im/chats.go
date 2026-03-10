package im

import (
	"fmt"
	"os"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/output"
)

func newChatsListCommand() *cobra.Command {
	var limit int
	var all bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "列举群聊",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, agent := cmdutil.ResolveTuple(cmd)
			_ = quiet

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

			type chatItem struct {
				ChatID string `json:"chat_id"`
				Name   string `json:"name"`
			}

			var items []chatItem
			var nextToken string
			pageToken := ""

			for {
				builder := larkim.NewListChatReqBuilder().
					PageSize(limit)
				if pageToken != "" {
					builder = builder.PageToken(pageToken)
				}

				resp, err := c.Client.Im.Chat.List(cmd.Context(), builder.Build(), c.RequestOptions()...)
				if err != nil {
					return fmt.Errorf("API_ERROR：%s", err.Error())
				}
				if !resp.Success() {
					return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
				}

				for _, chat := range resp.Data.Items {
					item := chatItem{}
					if chat.ChatId != nil {
						item.ChatID = *chat.ChatId
					}
					if chat.Name != nil {
						item.Name = *chat.Name
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

			if format == "json" {
				if agent {
					return output.PrintJSON(os.Stdout, PagedResponse{Items: items, NextCursor: nextToken})
				}
				return output.PrintJSON(os.Stdout, items)
			}

			for _, item := range items {
				fmt.Printf("%s\t%s\n", item.Name, item.ChatID)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "返回数量限制")
	cmd.Flags().BoolVar(&all, "all", false, "自动翻页获取全部")

	return cmd
}

func newChatsSearchCommand() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "search <keyword>",
		Short: "搜索群聊",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, _ := cmdutil.ResolveTuple(cmd)
			_ = quiet

			keyword := args[0]

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

			req := larkim.NewSearchChatReqBuilder().
				Query(keyword).
				PageSize(limit).
				Build()

			resp, err := c.Client.Im.Chat.Search(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			type chatItem struct {
				ChatID string `json:"chat_id"`
				Name   string `json:"name"`
			}

			var items []chatItem
			for _, chat := range resp.Data.Items {
				item := chatItem{}
				if chat.ChatId != nil {
					item.ChatID = *chat.ChatId
				}
				if chat.Name != nil {
					item.Name = *chat.Name
				}
				items = append(items, item)
			}

			if format == "json" {
				return output.PrintJSON(os.Stdout, items)
			}

			for _, item := range items {
				fmt.Printf("%s\t%s\n", item.Name, item.ChatID)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "返回数量限制")

	return cmd
}

// PagedResponse is used for agent mode paged responses.
type PagedResponse = cmdutil.PagedResponse
