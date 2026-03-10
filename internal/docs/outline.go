package docs

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/client"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/output"
)

func newOutlineCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outline <doc-url-or-token>",
		Short: "返回文档标题大纲",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, tokenMode, profile, cfg, domain, debug, quiet, _ := cmdutil.ResolveTuple(cmd)
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

			blocks, err := fetchAllBlocks(cmd, c, docToken)
			if err != nil {
				return err
			}

			outBlocks := convertBlocks(blocks)

			type HeadingItem struct {
				Level   int    `json:"level"`
				Text    string `json:"text"`
				BlockID string `json:"block_id"`
			}

			var headings []HeadingItem
			for _, b := range outBlocks {
				if b.BlockType >= 3 && b.BlockType <= 8 {
					level := b.BlockType - 2
					headings = append(headings, HeadingItem{
						Level:   level,
						Text:    b.TextContent(),
						BlockID: b.BlockID,
					})
				}
			}

			if format == "json" {
				return output.PrintJSON(os.Stdout, headings)
			}

			// text format: "H1  标题内容"
			for _, h := range headings {
				prefix := fmt.Sprintf("H%d", h.Level)
				indent := strings.Repeat("  ", h.Level-1)
				fmt.Printf("%s%s\t%s\n", indent, prefix, h.Text)
			}
			return nil
		},
	}

	return cmd
}
