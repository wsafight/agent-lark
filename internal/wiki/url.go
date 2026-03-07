package wiki

import (
	"net/url"
	"strings"
)

// ExtractWikiToken 从飞书 Wiki URL 解析出 space token 或 node token。
// 支持：https://company.feishu.cn/wiki/wikcnXXXXXX
func ExtractWikiToken(input string) string {
	if !strings.HasPrefix(input, "http") {
		return input
	}
	u, err := url.Parse(input)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	for i, p := range parts {
		if p == "wiki" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
