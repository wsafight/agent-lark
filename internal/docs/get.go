package docs

import (
	"fmt"
	"os"
	"strings"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
	"github.com/spf13/cobra"
	"github.com/wangshian/agent-lark/internal/client"
	"github.com/wangshian/agent-lark/internal/output"
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
				var jBlocks []jsonBlock
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
	ctx := cmd.Context()

	var allBlocks []*larkdocx.Block
	var pageToken string

	for {
		builder := larkdocx.NewListDocumentBlockReqBuilder().
			DocumentId(docToken).
			PageSize(200)
		if pageToken != "" {
			builder = builder.PageToken(pageToken)
		}

		resp, err := c.Client.Docx.DocumentBlock.List(
			ctx,
			builder.Build(),
			c.RequestOptions()...,
		)
		if err != nil {
			return nil, fmt.Errorf("API_ERROR：%s", err.Error())
		}
		if !resp.Success() {
			return nil, fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
		}

		allBlocks = append(allBlocks, resp.Data.Items...)

		hasMore := resp.Data.HasMore != nil && *resp.Data.HasMore
		if !hasMore {
			break
		}
		if resp.Data.PageToken == nil || *resp.Data.PageToken == "" {
			break
		}
		pageToken = *resp.Data.PageToken
	}

	return allBlocks, nil
}

// convertBlocks converts larkdocx blocks to output.BlockItem.
func convertBlocks(blocks []*larkdocx.Block) []*output.BlockItem {
	var result []*output.BlockItem
	for _, b := range blocks {
		if b == nil {
			continue
		}
		item := &output.BlockItem{}
		if b.BlockId != nil {
			item.BlockID = *b.BlockId
		}
		if b.BlockType != nil {
			item.BlockType = *b.BlockType
		}
		texts := extractTextsFromBlock(b)
		item.Texts = texts
		result = append(result, item)
	}
	return result
}

func extractTextsFromBlock(b *larkdocx.Block) []string {
	if b == nil || b.BlockType == nil {
		return nil
	}
	bt := *b.BlockType

	extractFromElements := func(elements []*larkdocx.TextElement) []string {
		var texts []string
		for _, el := range elements {
			if el == nil {
				continue
			}
			if el.TextRun != nil && el.TextRun.Content != nil {
				texts = append(texts, *el.TextRun.Content)
			}
		}
		return texts
	}

	switch bt {
	case 2: // text
		if b.Text != nil && b.Text.Elements != nil {
			return extractFromElements(b.Text.Elements)
		}
	case 3: // heading1
		if b.Heading1 != nil && b.Heading1.Elements != nil {
			return extractFromElements(b.Heading1.Elements)
		}
	case 4: // heading2
		if b.Heading2 != nil && b.Heading2.Elements != nil {
			return extractFromElements(b.Heading2.Elements)
		}
	case 5: // heading3
		if b.Heading3 != nil && b.Heading3.Elements != nil {
			return extractFromElements(b.Heading3.Elements)
		}
	case 6: // heading4
		if b.Heading4 != nil && b.Heading4.Elements != nil {
			return extractFromElements(b.Heading4.Elements)
		}
	case 7: // heading5
		if b.Heading5 != nil && b.Heading5.Elements != nil {
			return extractFromElements(b.Heading5.Elements)
		}
	case 8: // heading6
		if b.Heading6 != nil && b.Heading6.Elements != nil {
			return extractFromElements(b.Heading6.Elements)
		}
	case 9: // ordered_list
		if b.Ordered != nil && b.Ordered.Elements != nil {
			return extractFromElements(b.Ordered.Elements)
		}
	case 10: // bullet
		if b.Bullet != nil && b.Bullet.Elements != nil {
			return extractFromElements(b.Bullet.Elements)
		}
	case 11: // code
		if b.Code != nil && b.Code.Elements != nil {
			return extractFromElements(b.Code.Elements)
		}
	case 12: // quote
		if b.Quote != nil && b.Quote.Elements != nil {
			return extractFromElements(b.Quote.Elements)
		}
	}
	return nil
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
