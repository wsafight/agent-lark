package comments

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/docs"
	"github.com/wsafight/agent-lark/internal/output"
)

func newResolveCommand() *cobra.Command {
	var to string

	cmd := &cobra.Command{
		Use:   "resolve <文档URL>",
		Short: "标记评论为已解决",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.ResolveGlobalFlags(cmd)

			if to == "" {
				return fmt.Errorf("MISSING_FLAG：--to 为必填项（如 #1）")
			}

			indexStr := strings.TrimPrefix(to, "#")
			index, err := strconv.Atoi(indexStr)
			if err != nil || index < 1 {
				return fmt.Errorf("INVALID_FLAG：--to 格式应为 #N（如 #1）")
			}

			docURL := args[0]
			fileToken := docs.ExtractDocID(docURL)
			if fileToken == "" {
				return fmt.Errorf("INVALID_URL：无法从 %q 解析文档 token", docURL)
			}

			c, err := g.NewClient()
			if err != nil {
				return err
			}

			// First list comments to find the comment_id
			listReq := larkdrive.NewListFileCommentReqBuilder().
				FileType("docx").
				FileToken(fileToken).
				PageSize(200).
				Build()

			listResp, err := c.Client.Drive.FileComment.List(cmd.Context(), listReq, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !listResp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", listResp.Code, listResp.Msg)
			}

			comments := listResp.Data.Items
			if index > len(comments) {
				return fmt.Errorf("NOT_FOUND：评论 #%d 不存在，共有 %d 条评论", index, len(comments))
			}

			targetComment := comments[index-1]
			if targetComment.CommentId == nil {
				return fmt.Errorf("NOT_FOUND：评论 #%d 的 comment_id 为空", index)
			}
			commentID := *targetComment.CommentId

			// Patch comment to mark as solved
			patchReq := larkdrive.NewPatchFileCommentReqBuilder().
				FileType("docx").
				FileToken(fileToken).
				CommentId(commentID).
				Body(larkdrive.NewPatchFileCommentReqBodyBuilder().
					IsSolved(true).
					Build()).
				Build()

			patchResp, err := c.Client.Drive.FileComment.Patch(cmd.Context(), patchReq, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !patchResp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", patchResp.Code, patchResp.Msg)
			}

			type resolveResult struct {
				CommentID string `json:"comment_id"`
				IsSolved  bool   `json:"is_solved"`
			}

			result := resolveResult{CommentID: commentID, IsSolved: true}

			if g.Format == "json" {
				return output.PrintJSON(os.Stdout, result)
			}

			fmt.Printf("评论 #%d（%s）已标记为已解决\n", index, commentID)
			return nil
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "目标评论序号（如 #1）")
	return cmd
}
