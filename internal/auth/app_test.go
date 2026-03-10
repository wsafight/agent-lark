package auth

import "testing"

func TestNormalizeDomain(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   string
	}{
		{"https 前缀", "https://open.feishu.cn", "open.feishu.cn"},
		{"http 前缀", "http://open.feishu.cn", "open.feishu.cn"},
		{"尾部斜杠", "open.feishu.cn/", "open.feishu.cn"},
		{"带路径", "https://open.feishu.cn/open-apis/v3", "open.feishu.cn"},
		{"空字符串", "", ""},
		{"纯域名", "open.larksuite.com", "open.larksuite.com"},
		{"含前后空格", "  https://open.feishu.cn/  ", "open.feishu.cn"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeDomain(tc.input)
			if got != tc.want {
				t.Errorf("NormalizeDomain(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
