package task

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/wsafight/agent-lark/internal/cmdutil"
)

// PagedResponse is used for agent mode paged responses.
type PagedResponse = cmdutil.PagedResponse

func parseDueToMillis(input string) (string, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", nil
	}

	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		if len(s) <= 10 {
			return strconv.FormatInt(n*1000, 10), nil
		}
		return strconv.FormatInt(n, 10), nil
	}

	if t, err := time.Parse("2006-01-02", s); err == nil {
		return strconv.FormatInt(t.UnixMilli(), 10), nil
	}

	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return strconv.FormatInt(t.UnixMilli(), 10), nil
	}

	return "", fmt.Errorf("无法解析截止时间 %q，支持 YYYY-MM-DD / RFC3339 / Unix 时间戳（秒或毫秒）", input)
}

func normalizeTaskStatus(status string) (string, error) {
	s := strings.TrimSpace(strings.ToLower(status))
	switch s {
	case "", "todo", "done":
		return s, nil
	default:
		return "", fmt.Errorf("INVALID_STATUS：仅支持 todo|done")
	}
}

func deriveTaskStatus(apiStatus, completedAt string) string {
	if apiStatus != "" {
		return apiStatus
	}
	if completedAt != "" && completedAt != "0" {
		return "done"
	}
	return "todo"
}
