package docs

import (
	"fmt"
	"strings"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/output"
)

func newCreateCommand() *cobra.Command {
	var title string
	var folder string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "创建新文档",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.ResolveGlobalFlags(cmd)

			if title == "" {
				return fmt.Errorf("MISSING_FLAG：--title 是必填项")
			}

			c, err := g.NewClient()
			if err != nil {
				return err
			}

			builder := larkdocx.NewCreateDocumentReqBodyBuilder().
				Title(title)

			if folder != "" {
				folderToken := ExtractFolderToken(folder)
				if folderToken != "" {
					builder = builder.FolderToken(folderToken)
				}
			}

			req := larkdocx.NewCreateDocumentReqBuilder().
				Body(builder.Build()).
				Build()

			resp, err := c.Client.Docx.Document.Create(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			docToken := ""
			if resp.Data.Document != nil && resp.Data.Document.DocumentId != nil {
				docToken = *resp.Data.Document.DocumentId
			}

			effectiveOpenDomain := g.Domain
			if effectiveOpenDomain == "" && c.Cfg != nil {
				effectiveOpenDomain = c.Cfg.Domain
			}
			webDomain := "feishu.cn"
			if strings.Contains(strings.ToLower(effectiveOpenDomain), "larksuite") {
				webDomain = "larksuite.com"
			}

			docURL := fmt.Sprintf("https://%s/docx/%s", webDomain, docToken)
			output.PrintSuccess(g.Quiet, fmt.Sprintf("文档已创建：%s", docURL))

			if g.Agent {
				return output.PrintJSON(cmd.OutOrStdout(), map[string]string{
					"document_id": docToken,
					"url":         docURL,
				})
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "文档标题（必填）")
	cmd.Flags().StringVar(&folder, "folder", "", "目标文件夹 URL 或 token")

	return cmd
}
