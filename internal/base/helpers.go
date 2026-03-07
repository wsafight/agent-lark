package base

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/cmdutil"
)

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
type PagedResponse struct {
	Items      any    `json:"items"`
	NextCursor string `json:"next_cursor"`
}

func getGlobalFlags(cmd *cobra.Command) (format, tokenMode, profile, config, domain string, debug, quiet, agent bool) {
	g := cmdutil.GetGlobalFlags(cmd)
	return g.Format, g.TokenMode, g.Profile, g.Config, g.Domain, g.Debug, g.Quiet, g.Agent
}
