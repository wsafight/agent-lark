package docs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
	"github.com/spf13/cobra"
	"github.com/wangshian/agent-lark/internal/client"
	"github.com/wangshian/agent-lark/internal/output"
)

func newUpdateCommand() *cobra.Command {
	var content string
	var fromStdin bool
	var filePath string

	cmd := &cobra.Command{
		Use:   "update <doc-url-or-token>",
		Short: "向文档追加内容",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
			}

			docToken := ExtractDocID(args[0])
			if docToken == "" {
				return fmt.Errorf("INVALID_URL：无法解析文档 token")
			}

			// Determine content source
			var text string
			if fromStdin {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("STDIN_ERROR：%s", err.Error())
				}
				text = string(data)
			} else if filePath != "" {
				data, err := os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("FILE_ERROR：%s", err.Error())
				}
				text = string(data)
			} else if content != "" {
				text = content
			} else {
				return fmt.Errorf("MISSING_CONTENT：需要 --content、--stdin 或 --file 之一")
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

			// Parse content into paragraphs
			paragraphs := splitParagraphs(text)

			// Build paragraph blocks
			children := buildParagraphBlocks(paragraphs)

			req := larkdocx.NewCreateDocumentBlockChildrenReqBuilder().
				DocumentId(docToken).
				BlockId(docToken). // append to root block
				Body(
					larkdocx.NewCreateDocumentBlockChildrenReqBodyBuilder().
						Children(children).
						Index(-1). // append to end
						Build(),
				).
				Build()

			resp, err := c.Client.Docx.DocumentBlockChildren.Create(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			output.PrintSuccess(quiet, fmt.Sprintf("已追加 %d 段内容到文档 %s", len(paragraphs), docToken))

			if output.GlobalAgent {
				return output.PrintJSON(cmd.OutOrStdout(), map[string]any{
					"document_id":      docToken,
					"appended_blocks":  len(children),
				})
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&content, "content", "", "要追加的内容")
	cmd.Flags().BoolVar(&fromStdin, "stdin", false, "从标准输入读取内容")
	cmd.Flags().StringVar(&filePath, "file", "", "从文件读取内容")

	return cmd
}

func splitParagraphs(text string) []string {
	scanner := bufio.NewScanner(strings.NewReader(text))
	var paragraphs []string
	var current strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if current.Len() > 0 {
				paragraphs = append(paragraphs, current.String())
				current.Reset()
			}
		} else {
			if current.Len() > 0 {
				current.WriteString("\n")
			}
			current.WriteString(line)
		}
	}

	if current.Len() > 0 {
		paragraphs = append(paragraphs, current.String())
	}

	return paragraphs
}

func buildParagraphBlocks(paragraphs []string) []*larkdocx.Block {
	var blocks []*larkdocx.Block
	for _, p := range paragraphs {
		blockType := 2 // text paragraph
		content := p

		textRunContent := content
		textRun := larkdocx.NewTextRunBuilder().
			Content(textRunContent).
			Build()

		element := larkdocx.NewTextElementBuilder().
			TextRun(textRun).
			Build()

		textBlock := larkdocx.NewTextBuilder().
			Elements([]*larkdocx.TextElement{element}).
			Build()

		block := larkdocx.NewBlockBuilder().
			BlockType(blockType).
			Text(textBlock).
			Build()

		blocks = append(blocks, block)
	}
	return blocks
}
