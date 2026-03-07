package wiki

import "testing"

func TestExtractWikiToken(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "裸 token 直接返回",
			input: "wikcnABC123",
			want:  "wikcnABC123",
		},
		{
			name:  "标准 wiki URL",
			input: "https://company.feishu.cn/wiki/wikcnABC123",
			want:  "wikcnABC123",
		},
		{
			name:  "路径中包含 wiki 段也可识别",
			input: "https://company.larksuite.com/suite/wiki/wikcnXYZ789/node",
			want:  "wikcnXYZ789",
		},
		{
			name:  "缺失 wiki 段返回空",
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
			got := ExtractWikiToken(tc.input)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
