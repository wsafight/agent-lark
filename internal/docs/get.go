package docs

import (
	"fmt"
	"os"
	"strings"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/docxutil"
	"github.com/wsafight/agent-lark/internal/output"
)

func newGetCommand() *cobra.Command {
	var section string
	var metadataOnly bool

	cmd := &cobra.Command{
		Use:   "get <doc-url-or-token>",
		Short: "获取文档内容",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)
			_ = quiet

			docToken := ExtractDocID(args[0])
			if docToken == "" {
				return fmt.Errorf("INVALID_URL：无法解析文档 token")
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

			// 获取文档元数据
			docResp, err := c.Client.Docx.Document.Get(
				cmd.Context(),
				larkdocx.NewGetDocumentReqBuilder().DocumentId(docToken).Build(),
				c.RequestOptions()...,
			)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !docResp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", docResp.Code, docResp.Msg)
			}

			type DocMeta struct {
				DocumentID string `json:"document_id"`
				RevisionID int    `json:"revision_id"`
				Title      string `json:"title"`
			}

			meta := DocMeta{}
			if docResp.Data.Document != nil {
				doc := docResp.Data.Document
				if doc.DocumentId != nil {
					meta.DocumentID = *doc.DocumentId
				}
				if doc.RevisionId != nil {
					meta.RevisionID = *doc.RevisionId
				}
				if doc.Title != nil {
					meta.Title = *doc.Title
				}
			}

			if metadataOnly {
				if format == "json" {
					return output.PrintJSON(os.Stdout, meta)
				}
				fmt.Printf("ID:    %s\n", meta.DocumentID)
				fmt.Printf("Title: %s\n", meta.Title)
				fmt.Printf("Rev:   %d\n", meta.RevisionID)
				return nil
			}

			// 获取所有内容块（翻页）
			blocks, err := fetchAllBlocks(cmd, c, docToken)
			if err != nil {
				return err
			}

			// 转换为 output.BlockItem
			outBlocks := convertBlocks(blocks)

			// --section 过滤
			if section != "" {
				outBlocks = filterSection(outBlocks, section)
			}

			if format == "json" {
				type jsonBlock struct {
					BlockID   string `json:"block_id"`
					BlockType int    `json:"block_type"`
					Text      string `json:"text"`
				}
				jBlocks := make([]jsonBlock, 0, len(outBlocks))
				for _, b := range outBlocks {
					jBlocks = append(jBlocks, jsonBlock{
						BlockID:   b.BlockID,
						BlockType: b.BlockType,
						Text:      b.TextContent(),
					})
				}
				return output.PrintJSON(os.Stdout, map[string]any{
					"meta":   meta,
					"blocks": jBlocks,
				})
			}

			// 默认 md 格式
			md := output.BlocksToMarkdown(outBlocks)
			fmt.Print(md)
			return nil
		},
	}

	cmd.Flags().StringVar(&section, "section", "", "只返回含该关键词的标题章节内容")
	cmd.Flags().BoolVar(&metadataOnly, "metadata", false, "只输出元信息，不输出正文")

	return cmd
}

// fetchAllBlocks fetches all document blocks with pagination.
func fetchAllBlocks(cmd *cobra.Command, c *client.Result, docToken string) ([]*larkdocx.Block, error) {
	return docxutil.FetchAllBlocks(cmd.Context(), c, docToken)
}

// convertBlocks converts larkdocx blocks to output.BlockItem.
func convertBlocks(blocks []*larkdocx.Block) []*output.BlockItem {
	return docxutil.ConvertBlocks(blocks)
}

// filterSection returns blocks under a heading matching the keyword.
func filterSection(blocks []*output.BlockItem, keyword string) []*output.BlockItem {
	keyword = strings.ToLower(keyword)

	// Find the heading block that contains the keyword
	headingIdx := -1
	headingLevel := 0
	for i, b := range blocks {
		if b.BlockType >= 3 && b.BlockType <= 8 {
			text := strings.ToLower(b.TextContent())
			if strings.Contains(text, keyword) {
				headingIdx = i
				headingLevel = b.BlockType
				break
			}
		}
	}

	if headingIdx < 0 {
		return nil
	}

	result := []*output.BlockItem{blocks[headingIdx]}

	// Collect everything until the next heading of same or higher level
	for i := headingIdx + 1; i < len(blocks); i++ {
		b := blocks[i]
		if b.BlockType >= 3 && b.BlockType <= 8 && b.BlockType <= headingLevel {
			break
		}
		result = append(result, b)
	}

	return result
}
