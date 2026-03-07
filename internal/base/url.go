package base

import (
	"net/url"
	"strings"
)

// ParseBitableURL 从飞书多维表格 URL 解析出 app token 和 table ID。
// 支持：https://company.feishu.cn/base/bascnXXXXXX?table=tblYYYYYY
func ParseBitableURL(input string) (appToken, tableID string) {
	if !strings.HasPrefix(input, "http") {
		return input, ""
	}
	u, err := url.Parse(input)
	if err != nil {
		return "", ""
	}
	// 解析 app token
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	for i, p := range parts {
		if p == "base" && i+1 < len(parts) {
			appToken = parts[i+1]
			break
		}
	}
	// 解析 table ID
	tableID = u.Query().Get("table")
	return appToken, tableID
}
