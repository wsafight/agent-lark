package docs

import (
	"net/url"
	"strings"
)

// ExtractDocID 从飞书文档 URL 或直接 token 中解析出文档 token。
// 支持格式：
//   - https://company.feishu.cn/docx/doxcnXXXXXX
//   - https://company.larksuite.com/docx/doxcnXXXXXX
//   - doxcnXXXXXX（直接 token）
func ExtractDocID(input string) string {
	if !strings.HasPrefix(input, "http") {
		return input
	}
	u, err := url.Parse(input)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	for i, p := range parts {
		if p == "docx" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// ExtractFolderToken 从飞书文件夹 URL 中解析 token。
func ExtractFolderToken(input string) string {
	if !strings.HasPrefix(input, "http") {
		return input
	}
	u, err := url.Parse(input)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	for i, p := range parts {
		if (p == "folder" || p == "drive") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
