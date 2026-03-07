package perms

import "testing"

func TestExtractResourceToken(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantToken    string
		wantFileType string
	}{
		{
			name:         "裸 token 默认按 docx",
			input:        "doxcnABC123",
			wantToken:    "doxcnABC123",
			wantFileType: "docx",
		},
		{
			name:         "docx 链接",
			input:        "https://company.feishu.cn/docx/doxcnABC123",
			wantToken:    "doxcnABC123",
			wantFileType: "docx",
		},
		{
			name:         "wiki 链接",
			input:        "https://company.feishu.cn/wiki/wikcnABC123",
			wantToken:    "wikcnABC123",
			wantFileType: "wiki",
		},
		{
			name:         "base 链接映射为 bitable",
			input:        "https://company.feishu.cn/base/bascnABC123",
			wantToken:    "bascnABC123",
			wantFileType: "bitable",
		},
		{
			name:         "sheets 链接映射为 sheet",
			input:        "https://company.feishu.cn/sheets/shtcnABC123",
			wantToken:    "shtcnABC123",
			wantFileType: "sheet",
		},
		{
			name:         "file 链接",
			input:        "https://company.feishu.cn/file/filcnABC123",
			wantToken:    "filcnABC123",
			wantFileType: "file",
		},
		{
			name:         "folder 链接",
			input:        "https://company.feishu.cn/folder/fldcnABC123",
			wantToken:    "fldcnABC123",
			wantFileType: "folder",
		},
		{
			name:         "未知路径类型回退",
			input:        "https://company.feishu.cn/unknown/abc",
			wantToken:    "https://company.feishu.cn/unknown/abc",
			wantFileType: "docx",
		},
		{
			name:         "非法 URL 回退到原输入",
			input:        "http://%",
			wantToken:    "http://%",
			wantFileType: "docx",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotToken, gotFileType := ExtractResourceToken(tc.input)
			if gotToken != tc.wantToken {
				t.Errorf("token: got %q, want %q", gotToken, tc.wantToken)
			}
			if gotFileType != tc.wantFileType {
				t.Errorf("fileType: got %q, want %q", gotFileType, tc.wantFileType)
			}
		})
	}
}
