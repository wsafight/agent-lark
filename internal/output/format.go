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

// BlocksToMarkdown 将飞书文档 Block 列表转换为 Markdown 字符串。
func BlocksToMarkdown(blocks []*BlockItem) string {
	var sb strings.Builder
	for _, b := range blocks {
		switch b.BlockType {
		case 1: // page/root
			// skip
		case 2: // text
			sb.WriteString(b.TextContent())
			sb.WriteString("\n")
		case 3: // heading1
			sb.WriteString("# ")
			sb.WriteString(b.TextContent())
			sb.WriteString("\n")
		case 4: // heading2
			sb.WriteString("## ")
			sb.WriteString(b.TextContent())
			sb.WriteString("\n")
		case 5: // heading3
			sb.WriteString("### ")
			sb.WriteString(b.TextContent())
			sb.WriteString("\n")
		case 6: // heading4
			sb.WriteString("#### ")
			sb.WriteString(b.TextContent())
			sb.WriteString("\n")
		case 7: // heading5
			sb.WriteString("##### ")
			sb.WriteString(b.TextContent())
			sb.WriteString("\n")
		case 8: // heading6
			sb.WriteString("###### ")
			sb.WriteString(b.TextContent())
			sb.WriteString("\n")
		case 9: // ordered_list
			sb.WriteString("1. ")
			sb.WriteString(b.TextContent())
			sb.WriteString("\n")
		case 10: // bullet_list
			sb.WriteString("- ")
			sb.WriteString(b.TextContent())
			sb.WriteString("\n")
		case 11: // code
			sb.WriteString("```\n")
			sb.WriteString(b.TextContent())
			sb.WriteString("\n```\n")
		case 12: // quote
			sb.WriteString("> ")
			sb.WriteString(b.TextContent())
			sb.WriteString("\n")
		case 19: // divider
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
