package wiki

import (
	"fmt"
	"os"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
	larkwiki "github.com/larksuite/oapi-sdk-go/v3/service/wiki/v2"
	"github.com/spf13/cobra"
	"github.com/wangshian/agent-lark/internal/client"
	"github.com/wangshian/agent-lark/internal/output"
)

func newGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <wiki-url-or-token>",
		Short: "获取知识库页面内容",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, agent := getGlobalFlags(cmd)
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)
			_ = quiet

			wikiToken := ExtractWikiToken(args[0])
			if wikiToken == "" {
				return fmt.Errorf("INVALID_URL：无法解析 wiki token")
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

			// Get node to find obj_token (document token)
			nodeResp, err := c.Client.Wiki.Space.GetNode(
				cmd.Context(),
				larkwiki.NewGetNodeSpaceReqBuilder().
					Token(wikiToken).
					Build(),
				c.RequestOptions()...,
			)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !nodeResp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", nodeResp.Code, nodeResp.Msg)
			}

			objToken := ""
			if nodeResp.Data.Node != nil && nodeResp.Data.Node.ObjToken != nil {
				objToken = *nodeResp.Data.Node.ObjToken
			}
			if objToken == "" {
				return fmt.Errorf("API_ERROR：节点没有关联文档 token")
			}

			// Get document metadata
			docResp, err := c.Client.Docx.Document.Get(
				cmd.Context(),
				larkdocx.NewGetDocumentReqBuilder().DocumentId(objToken).Build(),
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

			// Get all blocks
			var allBlocks []*larkdocx.Block
			var pageToken string

			for {
				builder := larkdocx.NewListDocumentBlockReqBuilder().
					DocumentId(objToken).
					PageSize(200)
				if pageToken != "" {
					builder = builder.PageToken(pageToken)
				}

				blockResp, err := c.Client.Docx.DocumentBlock.List(
					cmd.Context(),
					builder.Build(),
					c.RequestOptions()...,
				)
				if err != nil {
					return fmt.Errorf("API_ERROR：%s", err.Error())
				}
				if !blockResp.Success() {
					return fmt.Errorf("API_ERROR：[%d] %s", blockResp.Code, blockResp.Msg)
				}

				allBlocks = append(allBlocks, blockResp.Data.Items...)

				hasMore := blockResp.Data.HasMore != nil && *blockResp.Data.HasMore
				if !hasMore {
					break
				}
				if blockResp.Data.PageToken == nil || *blockResp.Data.PageToken == "" {
					break
				}
				pageToken = *blockResp.Data.PageToken
			}

			outBlocks := convertBlocks(allBlocks)

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

			// Default: markdown
			md := output.BlocksToMarkdown(outBlocks)
			fmt.Print(md)
			return nil
		},
	}

	return cmd
}

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
	case 2:
		if b.Text != nil && b.Text.Elements != nil {
			return extractFromElements(b.Text.Elements)
		}
	case 3:
		if b.Heading1 != nil && b.Heading1.Elements != nil {
			return extractFromElements(b.Heading1.Elements)
		}
	case 4:
		if b.Heading2 != nil && b.Heading2.Elements != nil {
			return extractFromElements(b.Heading2.Elements)
		}
	case 5:
		if b.Heading3 != nil && b.Heading3.Elements != nil {
			return extractFromElements(b.Heading3.Elements)
		}
	case 6:
		if b.Heading4 != nil && b.Heading4.Elements != nil {
			return extractFromElements(b.Heading4.Elements)
		}
	case 7:
		if b.Heading5 != nil && b.Heading5.Elements != nil {
			return extractFromElements(b.Heading5.Elements)
		}
	case 8:
		if b.Heading6 != nil && b.Heading6.Elements != nil {
			return extractFromElements(b.Heading6.Elements)
		}
	case 9:
		if b.Ordered != nil && b.Ordered.Elements != nil {
			return extractFromElements(b.Ordered.Elements)
		}
	case 10:
		if b.Bullet != nil && b.Bullet.Elements != nil {
			return extractFromElements(b.Bullet.Elements)
		}
	case 11:
		if b.Code != nil && b.Code.Elements != nil {
			return extractFromElements(b.Code.Elements)
		}
	case 12:
		if b.Quote != nil && b.Quote.Elements != nil {
			return extractFromElements(b.Quote.Elements)
		}
	}
	return nil
}
