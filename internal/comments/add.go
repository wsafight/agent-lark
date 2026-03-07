package comments

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	"github.com/spf13/cobra"
	"github.com/wangshian/agent-lark/internal/client"
	"github.com/wangshian/agent-lark/internal/docs"
	"github.com/wangshian/agent-lark/internal/output"
)

type addResult struct {
	URL       string `json:"url"`
	Status    string `json:"status"`
	CommentID string `json:"comment_id,omitempty"`
	Error     string `json:"error,omitempty"`
	Message   string `json:"message,omitempty"`
}

type batchAddResponse struct {
	Total     int         `json:"total"`
	Succeeded int         `json:"succeeded"`
	Failed    int         `json:"failed"`
	Results   []addResult `json:"results"`
}

func newAddCommand() *cobra.Command {
	var content string
	var batch bool
	var onError string

	cmd := &cobra.Command{
		Use:   "add <文档URL>",
		Short: "添加文档评论",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)
			_ = quiet

			if content == "" {
				return fmt.Errorf("MISSING_FLAG：--content 为必填项")
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

			addComment := func(docURL string) addResult {
				fileToken := docs.ExtractDocID(docURL)
				if fileToken == "" {
					return addResult{URL: docURL, Status: "failed", Error: "INVALID_URL", Message: "无法解析文档 token"}
				}

				textRun := larkdrive.NewTextRunBuilder().Text(content).Build()
				element := larkdrive.NewReplyElementBuilder().TextRun(textRun).Build()
				commentContent := larkdrive.NewReplyContentBuilder().Elements([]*larkdrive.ReplyElement{element}).Build()
				reply := larkdrive.NewFileCommentReplyBuilder().Content(commentContent).Build()
				comment := larkdrive.NewFileCommentBuilder().
					ReplyList(larkdrive.NewReplyListBuilder().Replies([]*larkdrive.FileCommentReply{reply}).Build()).
					Build()

				req := larkdrive.NewCreateFileCommentReqBuilder().
					FileType("docx").
					FileToken(fileToken).
					FileComment(comment).
					Build()

				resp, err := c.Client.Drive.FileComment.Create(cmd.Context(), req, c.RequestOptions()...)
				if err != nil {
					return addResult{URL: docURL, Status: "failed", Error: "API_ERROR", Message: err.Error()}
				}
				if !resp.Success() {
					return addResult{URL: docURL, Status: "failed", Error: fmt.Sprintf("CODE_%d", resp.Code), Message: resp.Msg}
				}

				commentID := ""
				if resp.Data.CommentId != nil {
					commentID = *resp.Data.CommentId
				}
				return addResult{URL: docURL, Status: "succeeded", CommentID: commentID}
			}

			if !batch {
				if len(args) == 0 {
					return fmt.Errorf("MISSING_ARG：请提供文档 URL")
				}
				result := addComment(args[0])
				if result.Status == "failed" {
					return fmt.Errorf("%s：%s", result.Error, result.Message)
				}
				if format == "json" {
					return output.PrintJSON(os.Stdout, result)
				}
				fmt.Printf("评论已添加，comment_id: %s\n", result.CommentID)
				return nil
			}

			// Batch mode: read URLs from stdin
			var urls []string
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					urls = append(urls, line)
				}
			}

			var results []addResult
			succeeded := 0
			failed := 0
			var failedURLs []string

			for _, u := range urls {
				r := addComment(u)
				results = append(results, r)
				if r.Status == "succeeded" {
					succeeded++
				} else {
					failed++
					failedURLs = append(failedURLs, fmt.Sprintf("%s（%s）", u, r.Error))
					if onError == "stop" {
						break
					}
				}
			}

			batchResp := batchAddResponse{
				Total:     len(urls),
				Succeeded: succeeded,
				Failed:    failed,
				Results:   results,
			}

			if format == "json" {
				return output.PrintJSON(os.Stdout, batchResp)
			}

			fmt.Printf("✓ %d/%d 条评论已添加\n", succeeded, len(urls))
			if failed > 0 {
				fmt.Printf("✗ %d 条失败：%s\n", failed, strings.Join(failedURLs, ", "))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&content, "content", "", "评论内容（必填）")
	cmd.Flags().BoolVar(&batch, "batch", false, "批量模式，从 stdin 读取 URL 列表")
	cmd.Flags().StringVar(&onError, "on-error", "continue", "批量失败处理策略：continue|stop")
	return cmd
}
