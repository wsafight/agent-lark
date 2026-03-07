package perms

import (
	"net/url"
	"strings"
)

// ExtractResourceToken 从任意飞书资源 URL 提取 token 和 type。
// 支持 docx, wiki, base(bitable), sheet, folder, file 等类型。
func ExtractResourceToken(input string) (token, fileType string) {
	if !strings.HasPrefix(input, "http") {
		return input, "docx"
	}
	u, err := url.Parse(input)
	if err != nil {
		return input, "docx"
	}
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	for i, p := range parts {
		if i+1 >= len(parts) {
			break
		}
		switch p {
		case "docx":
			return parts[i+1], "docx"
		case "wiki":
			return parts[i+1], "wiki"
		case "base":
			return parts[i+1], "bitable"
		case "sheets":
			return parts[i+1], "sheet"
		case "file":
			return parts[i+1], "file"
		case "folder":
			return parts[i+1], "folder"
		}
	}
	return input, "docx"
}
