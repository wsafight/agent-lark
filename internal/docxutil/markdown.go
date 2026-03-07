package docxutil

import (
	"regexp"
	"strings"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
)

var orderedListRe = regexp.MustCompile(`^\d+\.\s+(.+)$`)

// MarkdownToBlocks parses a markdown string and returns a slice of Feishu document blocks.
// Supported syntax: # headings, - /* bullet lists, 1. ordered lists, > quotes, ``` code fences, --- dividers.
func MarkdownToBlocks(text string) []*larkdocx.Block {
	lines := strings.Split(text, "\n")
	var blocks []*larkdocx.Block
	inCode := false
	var codeLines []string

	for _, line := range lines {
		// code fence toggle
		if strings.HasPrefix(line, "```") {
			if inCode {
				blocks = append(blocks, makeTextBlock(11, strings.Join(codeLines, "\n")))
				codeLines = nil
				inCode = false
			} else {
				inCode = true
			}
			continue
		}
		if inCode {
			codeLines = append(codeLines, line)
			continue
		}

		stripped := strings.TrimRight(line, " ")

		switch {
		case stripped == "---" || stripped == "***" || stripped == "___":
			blocks = append(blocks, makeDivider())
		case strings.HasPrefix(stripped, "###### "):
			blocks = append(blocks, makeTextBlock(8, stripped[7:]))
		case strings.HasPrefix(stripped, "##### "):
			blocks = append(blocks, makeTextBlock(7, stripped[6:]))
		case strings.HasPrefix(stripped, "#### "):
			blocks = append(blocks, makeTextBlock(6, stripped[5:]))
		case strings.HasPrefix(stripped, "### "):
			blocks = append(blocks, makeTextBlock(5, stripped[4:]))
		case strings.HasPrefix(stripped, "## "):
			blocks = append(blocks, makeTextBlock(4, stripped[3:]))
		case strings.HasPrefix(stripped, "# "):
			blocks = append(blocks, makeTextBlock(3, stripped[2:]))
		case strings.HasPrefix(stripped, "> "):
			blocks = append(blocks, makeTextBlock(12, stripped[2:]))
		case strings.HasPrefix(stripped, "- ") || strings.HasPrefix(stripped, "* "):
			blocks = append(blocks, makeTextBlock(10, stripped[2:]))
		default:
			if m := orderedListRe.FindStringSubmatch(stripped); m != nil {
				blocks = append(blocks, makeTextBlock(9, m[1]))
			} else if stripped != "" {
				blocks = append(blocks, makeTextBlock(2, stripped))
			}
		}
	}

	// flush unclosed code block
	if inCode && len(codeLines) > 0 {
		blocks = append(blocks, makeTextBlock(11, strings.Join(codeLines, "\n")))
	}

	return blocks
}

// makeTextBlock builds a larkdocx.Block of the given type with text content.
// blockType: 2=text, 3-8=heading1-6, 9=ordered, 10=bullet, 11=code, 12=quote
func makeTextBlock(blockType int, content string) *larkdocx.Block {
	textBody := larkdocx.NewTextBuilder().
		Elements([]*larkdocx.TextElement{
			larkdocx.NewTextElementBuilder().
				TextRun(larkdocx.NewTextRunBuilder().Content(content).Build()).
				Build(),
		}).
		Build()

	b := larkdocx.NewBlockBuilder().BlockType(blockType)
	switch blockType {
	case 3:
		b = b.Heading1(textBody)
	case 4:
		b = b.Heading2(textBody)
	case 5:
		b = b.Heading3(textBody)
	case 6:
		b = b.Heading4(textBody)
	case 7:
		b = b.Heading5(textBody)
	case 8:
		b = b.Heading6(textBody)
	case 9:
		b = b.Ordered(textBody)
	case 10:
		b = b.Bullet(textBody)
	case 11:
		b = b.Code(textBody)
	case 12:
		b = b.Quote(textBody)
	default: // 2 and fallback
		b = b.Text(textBody)
	}
	return b.Build()
}

func makeDivider() *larkdocx.Block {
	return larkdocx.NewBlockBuilder().
		BlockType(19).
		Divider(larkdocx.NewDividerBuilder().Build()).
		Build()
}
