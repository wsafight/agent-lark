package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// --- FormatFromCmd ---

func TestFormatFromCmd(t *testing.T) {
	t.Run("非 agent 模式返回原始 format", func(t *testing.T) {
		GlobalAgent = false
		got := FormatFromCmd("table")
		if got != "table" {
			t.Errorf("got %q, want %q", got, "table")
		}
	})

	t.Run("agent 模式强制返回 json", func(t *testing.T) {
		GlobalAgent = true
		got := FormatFromCmd("table")
		if got != "json" {
			t.Errorf("got %q, want %q", got, "json")
		}
		GlobalAgent = false // 还原
	})
}

// --- PrintJSON ---

func TestPrintJSON(t *testing.T) {
	t.Run("输出带缩进的 JSON", func(t *testing.T) {
		var buf bytes.Buffer
		err := PrintJSON(&buf, map[string]string{"key": "value"})
		if err != nil {
			t.Fatalf("意外错误：%v", err)
		}
		var got map[string]string
		if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("输出不是合法 JSON：%v\n内容：%s", err, buf.String())
		}
		if got["key"] != "value" {
			t.Errorf("got[key]=%q, want %q", got["key"], "value")
		}
	})

	t.Run("输出 nil 为 null", func(t *testing.T) {
		var buf bytes.Buffer
		_ = PrintJSON(&buf, nil)
		if strings.TrimSpace(buf.String()) != "null" {
			t.Errorf("got %q, want %q", buf.String(), "null")
		}
	})
}

// --- PrintSuccess ---

func TestPrintSuccess(t *testing.T) {
	// PrintSuccess 仅向 stdout 打印；此处仅验证在不同模式下不 panic。
	t.Run("quiet=false agent=false 正常打印不 panic", func(t *testing.T) {
		GlobalAgent = false
		PrintSuccess(false, "ok")
	})

	t.Run("quiet=true 不打印不 panic", func(t *testing.T) {
		GlobalAgent = false
		PrintSuccess(true, "should not print")
	})

	t.Run("agent 模式不打印不 panic", func(t *testing.T) {
		GlobalAgent = true
		PrintSuccess(false, "should not print")
		GlobalAgent = false
	})
}

// --- BlocksToMarkdown ---

func TestBlocksToMarkdown(t *testing.T) {
	tests := []struct {
		name   string
		blocks []*BlockItem
		want   string
	}{
		{
			name: "普通文本块",
			blocks: []*BlockItem{
				{BlockType: 2, Texts: []string{"hello world"}},
			},
			want: "hello world\n",
		},
		{
			name: "heading1",
			blocks: []*BlockItem{
				{BlockType: 3, Texts: []string{"标题"}},
			},
			want: "# 标题\n",
		},
		{
			name: "heading2",
			blocks: []*BlockItem{
				{BlockType: 4, Texts: []string{"副标题"}},
			},
			want: "## 副标题\n",
		},
		{
			name: "heading3",
			blocks: []*BlockItem{
				{BlockType: 5, Texts: []string{"三级标题"}},
			},
			want: "### 三级标题\n",
		},
		{
			name: "有序列表",
			blocks: []*BlockItem{
				{BlockType: 9, Texts: []string{"item"}},
			},
			want: "1. item\n",
		},
		{
			name: "无序列表",
			blocks: []*BlockItem{
				{BlockType: 10, Texts: []string{"bullet"}},
			},
			want: "- bullet\n",
		},
		{
			name: "代码块",
			blocks: []*BlockItem{
				{BlockType: 11, Texts: []string{"fmt.Println()"}},
			},
			want: "```\nfmt.Println()\n```\n",
		},
		{
			name: "引用块",
			blocks: []*BlockItem{
				{BlockType: 12, Texts: []string{"引用文本"}},
			},
			want: "> 引用文本\n",
		},
		{
			name: "分割线",
			blocks: []*BlockItem{
				{BlockType: 19},
			},
			want: "---\n",
		},
		{
			name: "page 根节点被跳过",
			blocks: []*BlockItem{
				{BlockType: 1, Texts: []string{"忽略"}},
			},
			want: "",
		},
		{
			name: "未知类型，有文本则输出",
			blocks: []*BlockItem{
				{BlockType: 999, Texts: []string{"未知内容"}},
			},
			want: "未知内容\n",
		},
		{
			name: "未知类型，无文本则跳过",
			blocks: []*BlockItem{
				{BlockType: 999, Texts: []string{}},
			},
			want: "",
		},
		{
			name: "多段文本合并",
			blocks: []*BlockItem{
				{BlockType: 2, Texts: []string{"hello", " ", "world"}},
			},
			want: "hello world\n",
		},
		{
			name: "多块混合",
			blocks: []*BlockItem{
				{BlockType: 3, Texts: []string{"大标题"}},
				{BlockType: 2, Texts: []string{"正文"}},
				{BlockType: 19},
			},
			want: "# 大标题\n正文\n---\n",
		},
		{
			name: "空列表",
			blocks: nil,
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := BlocksToMarkdown(tc.blocks)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// --- BlockItem.TextContent ---

func TestBlockItemTextContent(t *testing.T) {
	t.Run("多段拼接", func(t *testing.T) {
		b := &BlockItem{Texts: []string{"a", "b", "c"}}
		if got := b.TextContent(); got != "abc" {
			t.Errorf("got %q, want %q", got, "abc")
		}
	})

	t.Run("空 Texts", func(t *testing.T) {
		b := &BlockItem{}
		if got := b.TextContent(); got != "" {
			t.Errorf("got %q, want %q", got, "")
		}
	})
}
