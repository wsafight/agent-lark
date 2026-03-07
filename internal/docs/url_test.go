package docs

import "testing"

func TestExtractDocID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "裸 token 直接返回",
			input: "doxcnABC123",
			want:  "doxcnABC123",
		},
		{
			name:  "标准 feishu 文档链接",
			input: "https://company.feishu.cn/docx/doxcnABC123",
			want:  "doxcnABC123",
		},
		{
			name:  "larksuite 文档链接带查询参数",
			input: "https://company.larksuite.com/docx/doxcnXYZ789?from=wiki",
			want:  "doxcnXYZ789",
		},
		{
			name:  "路径中包含 docx 段也可识别",
			input: "https://company.feishu.cn/suite/docx/doxcnABC123/view",
			want:  "doxcnABC123",
		},
		{
			name:  "缺失 docx 段返回空",
			input: "https://company.feishu.cn/wiki/wikcnABC123",
			want:  "",
		},
		{
			name:  "非法 URL 返回空",
			input: "http://%",
			want:  "",
		},
		{
			name:  "空字符串返回空",
			input: "",
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ExtractDocID(tc.input)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestExtractFolderToken(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "裸 folder token 直接返回",
			input: "fldcnABC123",
			want:  "fldcnABC123",
		},
		{
			name:  "folder 链接提取 token",
			input: "https://company.feishu.cn/folder/fldcnABC123",
			want:  "fldcnABC123",
		},
		{
			name:  "drive 链接提取 token",
			input: "https://company.feishu.cn/drive/fldcnXYZ789",
			want:  "fldcnXYZ789",
		},
		{
			name:  "缺失 folder 和 drive 段返回空",
			input: "https://company.feishu.cn/docx/doxcnABC123",
			want:  "",
		},
		{
			name:  "非法 URL 返回空",
			input: "http://%",
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ExtractFolderToken(tc.input)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
