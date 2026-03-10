package comments

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/docs"
	"github.com/wsafight/agent-lark/internal/output"
)

func newReplyCommand() *cobra.Command {
	var to string
	var content string

	cmd := &cobra.Command{
		Use:   "reply <文档URL>",
		Short: "回复文档评论",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, _ := cmdutil.ResolveTuple(cmd)
			_ = quiet

			if to == "" {
				return fmt.Errorf("MISSING_FLAG：--to 为必填项（如 #1）")
			}
			if content == "" {
				return fmt.Errorf("MISSING_FLAG：--content 为必填项")
			}

			indexStr := strings.TrimPrefix(to, "#")
			_, err := strconv.Atoi(indexStr)
			if err != nil {
				return fmt.Errorf("INVALID_FLAG：--to 格式应为 #N（如 #1）")
			}

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

			// SDK v3.5.3 does not support FileCommentReply.Create;
			// create a new top-level comment as a workaround.
			replyText := fmt.Sprintf("[回复 %s] %s", to, content)
			textRun := larkdrive.NewTextRunBuilder().Text(replyText).Build()
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
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			type replyResult struct {
				CommentID string `json:"comment_id"`
				To        string `json:"to"`
			}

			result := replyResult{To: to}
			if resp.Data.CommentId != nil {
				result.CommentID = *resp.Data.CommentId
			}

			if format == "json" {
				return output.PrintJSON(os.Stdout, result)
			}

			fmt.Printf("回复已添加，comment_id: %s\n", result.CommentID)
			return nil
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "目标评论序号（如 #1）")
	cmd.Flags().StringVar(&content, "content", "", "回复内容（必填）")
	return cmd
}
