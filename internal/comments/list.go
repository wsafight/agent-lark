package comments

import (
	"fmt"
	"os"
	"time"

	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/docs"
	"github.com/wsafight/agent-lark/internal/output"
)

type commentItem struct {
	Index     int    `json:"index"`
	CommentID string `json:"comment_id"`
	Author    string `json:"author"`
	Date      string `json:"date"`
	Content   string `json:"content"`
	IsSolved  bool   `json:"is_solved"`
}

func newListCommand() *cobra.Command {
	var limit int
	var all bool
	var cursor string

	cmd := &cobra.Command{
		Use:   "list <文档URL>",
		Short: "列举文档评论",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, agent := cmdutil.ResolveTuple(cmd)
			_ = quiet

			docURL := args[0]
			fileToken := docs.ExtractDocID(docURL)
			if fileToken == "" {
				return fmt.Errorf("INVALID_URL：无法从 %q 解析文档 token", docURL)
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

			var items []commentItem
			var nextToken string

			pageSize := int32(limit)
			if pageSize <= 0 {
				pageSize = 50
			}
			pageToken := cursor
			index := 1

			for {
				reqBuilder := larkdrive.NewListFileCommentReqBuilder().
					FileType("docx").
					FileToken(fileToken).
					PageSize(int(pageSize))
				if pageToken != "" {
					reqBuilder = reqBuilder.PageToken(pageToken)
				}

				resp, err := c.Client.Drive.FileComment.List(cmd.Context(), reqBuilder.Build(), c.RequestOptions()...)
				if err != nil {
					return fmt.Errorf("API_ERROR：%s", err.Error())
				}
				if !resp.Success() {
					return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
				}

				for _, comment := range resp.Data.Items {
					item := commentItem{Index: index}
					index++
					if comment.CommentId != nil {
						item.CommentID = *comment.CommentId
					}
					if comment.UserId != nil {
						item.Author = *comment.UserId
					}
					if comment.CreateTime != nil {
						t := time.Unix(int64(*comment.CreateTime)/1000, 0)
						item.Date = t.Format("2006-01-02 15:04")
					}
					if comment.IsSolved != nil {
						item.IsSolved = *comment.IsSolved
					}
					// Extract text content from reply list
					if len(comment.ReplyList.Replies) > 0 {
						first := comment.ReplyList.Replies[0]
						if first.Content != nil && len(first.Content.Elements) > 0 {
							for _, el := range first.Content.Elements {
								if el.TextRun != nil && el.TextRun.Text != nil {
									item.Content += *el.TextRun.Text
								}
							}
						}
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
				status := "未解决"
				if item.IsSolved {
					status = "已解决"
				}
				fmt.Printf("#%d  %s  %s  %s  [%s]\n", item.Index, item.Author, item.Date, item.Content, status)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 50, "返回数量限制")
	cmd.Flags().BoolVar(&all, "all", false, "自动翻页获取全部")
	cmd.Flags().StringVar(&cursor, "cursor", "", "分页游标")
	return cmd
}
