package base

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wsafight/agent-lark/internal/cmdutil"
)

// parseBitableURLStrict 解析多维表格 URL，返回 appToken 和 tableID，失败时返回带错误码的 error。
func parseBitableURLStrict(rawURL string) (appToken, tableID string, err error) {
	appToken, tableID = ParseBitableURL(rawURL)
	if appToken == "" {
		return "", "", fmt.Errorf("INVALID_URL：无法解析多维表格 URL")
	}
	if tableID == "" {
		return "", "", fmt.Errorf("INVALID_URL：URL 中缺少 table 参数")
	}
	return
}

// parseFieldPair 解析 "key=value" 格式的字段对。
// value 若能解析为 JSON 原始类型（数字、布尔、null），则使用对应 Go 类型；否则作为字符串。
func parseFieldPair(pair string) (string, interface{}, error) {
	idx := strings.Index(pair, "=")
	if idx < 0 {
		return "", nil, fmt.Errorf("格式错误 %q，期望 key=value", pair)
	}
	key := pair[:idx]
	val := pair[idx+1:]
	if key == "" {
		return "", nil, fmt.Errorf("格式错误 %q：key 不能为空", pair)
	}
	var parsed interface{}
	if err := json.Unmarshal([]byte(val), &parsed); err == nil {
		return key, parsed, nil
	}
	return key, val, nil
}

// PagedResponse is used for agent mode paged responses.
type PagedResponse = cmdutil.PagedResponse
