package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// GlobalAgent は --agent フラグの値を保持する（main.go で設定）
var GlobalAgent bool

// PrintJSON は任意の値を JSON として stdout に出力する。
func PrintJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// PrintError はエラーを適切な形式で stderr に出力する。
func PrintError(err error) {
	PrintErrorCode("ERROR", err.Error(), "")
}

// PrintErrorCode は構造化エラーを stderr に出力する。
func PrintErrorCode(code, message, hint string) {
	if GlobalAgent {
		v := map[string]string{"error": code, "message": message}
		if hint != "" {
			v["hint"] = hint
		}
		enc := json.NewEncoder(os.Stderr)
		_ = enc.Encode(v)
	} else {
		if hint != "" {
			fmt.Fprintf(os.Stderr, "[ERROR] %s: %s\n  建议：%s\n", code, message, hint)
		} else {
			fmt.Fprintf(os.Stderr, "[ERROR] %s: %s\n", code, message)
		}
	}
}

// FormatFromCmd 从命令获取格式标志，考虑 --agent 模式。
func FormatFromCmd(format string) string {
	if GlobalAgent {
		return "json"
	}
	return format
}

// PrintSuccess 打印成功消息（quiet 模式下跳过）。
func PrintSuccess(quiet bool, msg string) {
	if !quiet && !GlobalAgent {
		fmt.Println(msg)
	}
}

// Block type constants for Feishu docx blocks.
const (
	BlockTypePage        = 1
	BlockTypeText        = 2
	BlockTypeHeading1    = 3
	BlockTypeHeading2    = 4
	BlockTypeHeading3    = 5
	BlockTypeHeading4    = 6
	BlockTypeHeading5    = 7
	BlockTypeHeading6    = 8
	BlockTypeOrderedList = 9
	BlockTypeBulletList  = 10
	BlockTypeCode        = 11
	BlockTypeQuote       = 12
	BlockTypeDivider     = 19
)

// BlocksToMarkdown 将飞书文档 Block 列表转换为 Markdown 字符串。
func BlocksToMarkdown(blocks []*BlockItem) string {
	var sb strings.Builder
	orderedCounter := 0
	for _, b := range blocks {
		if b.BlockType != BlockTypeOrderedList {
			orderedCounter = 0
		}
		switch b.BlockType {
		case BlockTypePage: // skip root block
		case BlockTypeText:
			sb.WriteString(b.TextContent())
			sb.WriteString("\n")
		case BlockTypeHeading1, BlockTypeHeading2, BlockTypeHeading3,
			BlockTypeHeading4, BlockTypeHeading5, BlockTypeHeading6:
			level := b.BlockType - BlockTypeHeading1 + 1
			sb.WriteString(strings.Repeat("#", level))
			sb.WriteString(" ")
			sb.WriteString(b.TextContent())
			sb.WriteString("\n")
		case BlockTypeOrderedList:
			orderedCounter++
			fmt.Fprintf(&sb, "%d. %s\n", orderedCounter, b.TextContent())
		case BlockTypeBulletList:
			sb.WriteString("- ")
			sb.WriteString(b.TextContent())
			sb.WriteString("\n")
		case BlockTypeCode:
			sb.WriteString("```\n")
			sb.WriteString(b.TextContent())
			sb.WriteString("\n```\n")
		case BlockTypeQuote:
			sb.WriteString("> ")
			sb.WriteString(b.TextContent())
			sb.WriteString("\n")
		case BlockTypeDivider:
			sb.WriteString("---\n")
		default:
			text := b.TextContent()
			if text != "" {
				sb.WriteString(text)
				sb.WriteString("\n")
			}
		}
	}
	return sb.String()
}

// BlockItem 是简化的 Block 表示，用于输出渲染。
type BlockItem struct {
	BlockID   string
	BlockType int
	Texts     []string
}

func (b *BlockItem) TextContent() string {
	return strings.Join(b.Texts, "")
}
