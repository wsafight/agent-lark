package docxutil

import (
	"testing"
)

func TestMarkdownToBlocks(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantTypes  []int // 期望的 BlockType 序列
		wantLength int
	}{
		{
			name:      "空字符串",
			input:     "",
			wantTypes: nil,
		},
		{
			name:      "普通段落",
			input:     "hello world",
			wantTypes: []int{2},
		},
		{
			name:      "一级标题",
			input:     "# 大标题",
			wantTypes: []int{3},
		},
		{
			name:      "二级标题",
			input:     "## 副标题",
			wantTypes: []int{4},
		},
		{
			name:      "三级标题",
			input:     "### 三级",
			wantTypes: []int{5},
		},
		{
			name:      "四级标题",
			input:     "#### 四级",
			wantTypes: []int{6},
		},
		{
			name:      "五级标题",
			input:     "##### 五级",
			wantTypes: []int{7},
		},
		{
			name:      "六级标题",
			input:     "###### 六级",
			wantTypes: []int{8},
		},
		{
			name:      "无序列表 -",
			input:     "- 项目",
			wantTypes: []int{10},
		},
		{
			name:      "无序列表 *",
			input:     "* 项目",
			wantTypes: []int{10},
		},
		{
			name:      "有序列表",
			input:     "1. 第一项",
			wantTypes: []int{9},
		},
		{
			name:      "多个有序列表项",
			input:     "1. 第一项\n2. 第二项\n3. 第三项",
			wantTypes: []int{9, 9, 9},
		},
		{
			name:      "引用块",
			input:     "> 引用文字",
			wantTypes: []int{12},
		},
		{
			name:      "分割线 ---",
			input:     "---",
			wantTypes: []int{19},
		},
		{
			name:      "分割线 ***",
			input:     "***",
			wantTypes: []int{19},
		},
		{
			name:      "分割线 ___",
			input:     "___",
			wantTypes: []int{19},
		},
		{
			name:      "代码块",
			input:     "```\nfmt.Println()\n```",
			wantTypes: []int{11},
		},
		{
			name:      "未闭合代码块仍输出",
			input:     "```\nsome code",
			wantTypes: []int{11},
		},
		{
			name:      "空行被忽略",
			input:     "第一段\n\n第二段",
			wantTypes: []int{2, 2},
		},
		{
			name:      "混合内容",
			input:     "# 标题\n- 列表\n> 引用\n普通文本\n---",
			wantTypes: []int{3, 10, 12, 2, 19},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			blocks := MarkdownToBlocks(tc.input)
			if len(blocks) != len(tc.wantTypes) {
				types := make([]int, len(blocks))
				for i, b := range blocks {
					if b != nil && b.BlockType != nil {
						types[i] = *b.BlockType
					}
				}
				t.Fatalf("块数量：got %d, want %d\n输入：%q\n块类型：%v",
					len(blocks), len(tc.wantTypes), tc.input, types)
			}
			for i, b := range blocks {
				if b == nil || b.BlockType == nil {
					t.Errorf("block[%d] 为 nil 或 BlockType 为 nil", i)
					continue
				}
				if *b.BlockType != tc.wantTypes[i] {
					t.Errorf("block[%d] BlockType：got %d, want %d", i, *b.BlockType, tc.wantTypes[i])
				}
			}
		})
	}
}

// TestMarkdownToBlocksCodeContent 验证代码块内容被正确捕获（不被解析为其他类型）
func TestMarkdownToBlocksCodeContent(t *testing.T) {
	input := "```\n# 这是代码注释，不是标题\n- 也不是列表\n```"
	blocks := MarkdownToBlocks(input)
	if len(blocks) != 1 {
		t.Fatalf("期望 1 个块，got %d", len(blocks))
	}
	if blocks[0].BlockType == nil || *blocks[0].BlockType != 11 {
		t.Errorf("期望 BlockType=11（代码块），got %v", blocks[0].BlockType)
	}
}
