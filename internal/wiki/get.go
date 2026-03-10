package wiki

import (
	"fmt"
	"os"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
	larkwiki "github.com/larksuite/oapi-sdk-go/v3/service/wiki/v2"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/docxutil"
	"github.com/wsafight/agent-lark/internal/output"
)

func newGetCommand() *cobra.Command {
	var contentBoundaries bool
	var maxChars int

	cmd := &cobra.Command{
		Use:   "get <wiki-url-or-token>",
		Short: "获取知识库页面内容",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.ResolveGlobalFlags(cmd)

			wikiToken := ExtractWikiToken(args[0])
			if wikiToken == "" {
				return fmt.Errorf("INVALID_URL：无法解析 wiki token")
			}

			c, err := g.NewClient()
			if err != nil {
				return err
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

			allBlocks, err := docxutil.FetchAllBlocks(cmd.Context(), c, objToken)
			if err != nil {
				return err
			}
			outBlocks := docxutil.ConvertBlocks(allBlocks)

			if g.Format == "json" {
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

			// Default: markdown
			md := output.BlocksToMarkdown(outBlocks)
			if maxChars > 0 && len(md) > maxChars {
				md = md[:maxChars] + "\n…[已截断]"
			}
			if contentBoundaries {
				fmt.Printf("<document source=\"feishu://wiki/%s\">\n%s\n</document>\n", objToken, md)
			} else {
				fmt.Print(md)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&contentBoundaries, "content-boundaries", false, "用 <document> 标签包裹输出，防止提示注入")
	cmd.Flags().IntVar(&maxChars, "max-chars", 0, "截断输出，最多输出 N 个字符（0 表示不限制）")
	return cmd
}
