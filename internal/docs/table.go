package docs

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/output"
)

func newTableCommand() *cobra.Command {
	var after string
	var before string
	var matchIndex int
	var rows int
	var cols int
	var headers bool
	var dataFile string
	var dataJSON string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "table <doc-url-or-token>",
		Short: "在文档中插入表格",
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

			// Get all blocks
			rawBlocks, err := fetchAllBlocks(cmd, c, docToken)
			if err != nil {
				return err
			}

			outBlocks := convertBlocks(rawBlocks)

			// Find insertion point
			keyword := after
			if keyword == "" {
				keyword = before
			}

			type Match struct {
				Index   int    `json:"index"`
				BlockID string `json:"block_id"`
				Text    string `json:"text"`
			}

			var matches []Match
			if keyword != "" {
				kw := strings.ToLower(keyword)
				for i, b := range outBlocks {
					content := b.TextContent()
					text := strings.ToLower(content)
					if strings.Contains(text, kw) {
						matches = append(matches, Match{
							Index:   i,
							BlockID: b.BlockID,
							Text:    content,
						})
					}
				}

				if len(matches) > 1 && matchIndex < 0 {
					if format == "json" {
						_ = output.PrintJSON(os.Stdout, map[string]any{
							"error":   "AMBIGUOUS_MATCH",
							"message": fmt.Sprintf("找到 %d 个匹配段落，请用 --match-index 指定", len(matches)),
							"matches": matches,
						})
						return fmt.Errorf("AMBIGUOUS_MATCH：找到 %d 个匹配段落，请用 --match-index 指定", len(matches))
					}
					fmt.Fprintf(os.Stderr, "[ERROR] AMBIGUOUS_MATCH: 找到 %d 个匹配段落，请用 --match-index 指定\n", len(matches))
					for i, m := range matches {
						fmt.Fprintf(os.Stderr, "  [%d] block_id=%s text=%q\n", i, m.BlockID, m.Text)
					}
					return fmt.Errorf("AMBIGUOUS_MATCH：找到 %d 个匹配段落，请用 --match-index 指定", len(matches))
				}

				if len(matches) == 0 {
					return fmt.Errorf("NOT_FOUND：未找到包含 %q 的段落", keyword)
				}
			}

			// Determine insert position
			insertBlockID := docToken // default: append to root
			insertIndex := -1

			if len(matches) > 0 {
				selectedMatchIdx := 0
				if matchIndex >= 0 {
					selectedMatchIdx = matchIndex
				}
				if selectedMatchIdx >= len(matches) {
					return fmt.Errorf("INVALID_INDEX：--match-index %d 超出范围（共 %d 个匹配）", selectedMatchIdx, len(matches))
				}

				selectedMatch := matches[selectedMatchIdx]
				insertBlockID = docToken // insert in root block

				if after != "" {
					// insert after the matched block: index is position + 1
					insertIndex = selectedMatch.Index + 1
				} else {
					// insert before the matched block
					insertIndex = selectedMatch.Index
				}
			}

			// Load CSV data if provided
			var tableData [][]string
			if dataFile != "" {
				f, err := os.Open(dataFile)
				if err != nil {
					return fmt.Errorf("FILE_ERROR：%s", err.Error())
				}
				defer f.Close()
				r := csv.NewReader(f)
				tableData, err = r.ReadAll()
				if err != nil {
					return fmt.Errorf("CSV_ERROR：%s", err.Error())
				}
			} else if dataJSON != "" {
				if err := json.Unmarshal([]byte(dataJSON), &tableData); err != nil {
					return fmt.Errorf("JSON_ERROR：%s", err.Error())
				}
			}

			// Determine rows/cols from data
			if len(tableData) > 0 {
				rows = len(tableData)
				if len(tableData[0]) > cols {
					cols = len(tableData[0])
				}
			}
			if headers && len(tableData) == 0 {
				rows++ // add a header row
			}
			if rows <= 0 || cols <= 0 {
				return fmt.Errorf("INVALID_TABLE_SIZE：行列必须大于 0")
			}

			if dryRun {
				preview := map[string]any{
					"action":          "insert_table",
					"document_id":     docToken,
					"insert_after":    after,
					"insert_before":   before,
					"insert_block_id": insertBlockID,
					"insert_index":    insertIndex,
					"rows":            rows,
					"cols":            cols,
					"headers":         headers,
				}
				if len(tableData) > 0 {
					preview["data"] = tableData
				}
				fmt.Println("[DRY RUN] 预览操作：")
				return output.PrintJSON(os.Stdout, preview)
			}

			// Build table block
			tableBlockType := 31
			table := larkdocx.NewTableBuilder().
				Property(
					larkdocx.NewTablePropertyBuilder().
						RowSize(rows).
						ColumnSize(cols).
						HeaderRow(headers).
						Build(),
				).
				Build()
			block := larkdocx.NewBlockBuilder().
				BlockType(tableBlockType).
				Table(table).
				Build()

			reqBody := larkdocx.NewCreateDocumentBlockChildrenReqBodyBuilder().
				Children([]*larkdocx.Block{block}).
				Index(insertIndex).
				Build()

			req := larkdocx.NewCreateDocumentBlockChildrenReqBuilder().
				DocumentId(docToken).
				BlockId(insertBlockID).
				Body(reqBody).
				Build()

			resp, err := c.Client.Docx.DocumentBlockChildren.Create(cmd.Context(), req, c.RequestOptions()...)
			if err != nil {
				return fmt.Errorf("API_ERROR：%s", err.Error())
			}
			if !resp.Success() {
				return fmt.Errorf("API_ERROR：[%d] %s", resp.Code, resp.Msg)
			}

			// Fill table cell text when data is provided.
			if len(tableData) > 0 && len(resp.Data.Children) > 0 {
				tableBlock := resp.Data.Children[0]
				if tableBlock != nil && tableBlock.Table != nil && len(tableBlock.Table.Cells) > 0 {
					for r := 0; r < rows && r < len(tableData); r++ {
						row := tableData[r]
						for cidx := 0; cidx < cols && cidx < len(row); cidx++ {
							text := strings.TrimSpace(row[cidx])
							if text == "" {
								continue
							}
							cellIndex := r*cols + cidx
							if cellIndex >= len(tableBlock.Table.Cells) {
								continue
							}
							cellID := tableBlock.Table.Cells[cellIndex]
							if err := appendTextToTableCell(cmd, c, docToken, cellID, text); err != nil {
								return err
							}
						}
					}
				}
			}

			output.PrintSuccess(quiet, fmt.Sprintf("表格已插入文档 %s（%d 行 × %d 列）", docToken, rows, cols))

			if output.GlobalAgent {
				return output.PrintJSON(cmd.OutOrStdout(), map[string]any{
					"document_id": docToken,
					"rows":        rows,
					"cols":        cols,
				})
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&after, "after", "", "在含此关键词的段落后插入")
	cmd.Flags().StringVar(&before, "before", "", "在含此关键词的段落前插入")
	cmd.Flags().IntVar(&matchIndex, "match-index", -1, "当有多个匹配时，指定使用第几个（0起始）")
	cmd.Flags().IntVar(&rows, "rows", 3, "表格行数")
	cmd.Flags().IntVar(&cols, "cols", 3, "表格列数")
	cmd.Flags().BoolVar(&headers, "headers", false, "第一行作为表头")
	cmd.Flags().StringVar(&dataFile, "data", "", "CSV 数据文件路径")
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "JSON 格式表格数据")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "只预览，不实际写入")

	return cmd
}

func appendTextToTableCell(cmd *cobra.Command, c *client.Result, docToken, cellID, text string) error {
	textRun := larkdocx.NewTextRunBuilder().
		Content(text).
		Build()
	element := larkdocx.NewTextElementBuilder().
		TextRun(textRun).
		Build()
	textBlock := larkdocx.NewTextBuilder().
		Elements([]*larkdocx.TextElement{element}).
		Build()
	block := larkdocx.NewBlockBuilder().
		BlockType(2).
		Text(textBlock).
		Build()

	req := larkdocx.NewCreateDocumentBlockChildrenReqBuilder().
		DocumentId(docToken).
		BlockId(cellID).
		Body(
			larkdocx.NewCreateDocumentBlockChildrenReqBodyBuilder().
				Children([]*larkdocx.Block{block}).
				Index(-1).
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
	return nil
}
