package task

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// --- parseDueToMillis ---

func TestParseDueToMillis(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, got string)
	}{
		{
			name:  "空字符串返回空",
			input: "",
			check: func(t *testing.T, got string) {
				if got != "" {
					t.Errorf("got %q, want empty", got)
				}
			},
		},
		{
			name:  "Unix 秒时间戳转毫秒",
			input: "1700000000",
			check: func(t *testing.T, got string) {
				if got != "1700000000000" {
					t.Errorf("got %q, want %q", got, "1700000000000")
				}
			},
		},
		{
			name:  "Unix 毫秒时间戳保持不变",
			input: "1700000000000",
			check: func(t *testing.T, got string) {
				if got != "1700000000000" {
					t.Errorf("got %q, want %q", got, "1700000000000")
				}
			},
		},
		{
			name:  "YYYY-MM-DD 日期",
			input: "2024-11-15",
			check: func(t *testing.T, got string) {
				ms, err := strconv.ParseInt(got, 10, 64)
				if err != nil {
					t.Fatalf("结果不是数字：%v", err)
				}
				ts := time.UnixMilli(ms)
				if ts.Year() != 2024 || ts.Month() != 11 || ts.Day() != 15 {
					t.Errorf("日期不匹配：got %v", ts)
				}
			},
		},
		{
			name:  "RFC3339 时间",
			input: "2024-11-15T10:00:00Z",
			check: func(t *testing.T, got string) {
				ms, err := strconv.ParseInt(got, 10, 64)
				if err != nil {
					t.Fatalf("结果不是数字：%v", err)
				}
				ts := time.UnixMilli(ms)
				if ts.UTC().Year() != 2024 || ts.UTC().Month() != 11 || ts.UTC().Day() != 15 {
					t.Errorf("日期不匹配：got %v", ts)
				}
			},
		},
		{
			name:    "无效格式",
			input:   "next-friday",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseDueToMillis(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("期望错误，但没有报错，got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("意外错误：%v", err)
			}
			tc.check(t, got)
		})
	}
}

// --- normalizeTaskStatus ---

func TestNormalizeTaskStatus(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"todo", "todo", false},
		{"done", "done", false},
		{"TODO", "todo", false},  // 大小写不敏感
		{"DONE", "done", false},
		{"  done  ", "done", false}, // 去除空白
		{"", "", false},
		{"in_progress", "", true},
		{"pending", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := normalizeTaskStatus(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("期望错误，但没有报错，got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("意外错误：%v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// --- deriveTaskStatus ---

func TestDeriveTaskStatus(t *testing.T) {
	tests := []struct {
		name        string
		apiStatus   string
		completedAt string
		want        string
	}{
		{"api 状态优先", "done", "1700000000000", "done"},
		{"api 状态为 todo 时优先", "todo", "1700000000000", "todo"},
		{"completedAt 非零推断为 done", "", "1700000000000", "done"},
		{"completedAt 为 0 推断为 todo", "", "0", "todo"},
		{"completedAt 为空推断为 todo", "", "", "todo"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := deriveTaskStatus(tc.apiStatus, tc.completedAt)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// --- parseDueToMillis 边界：前导/尾随空格 ---

func TestParseDueToMillisTrimSpace(t *testing.T) {
	got, err := parseDueToMillis("  2024-01-01  ")
	if err != nil {
		t.Fatalf("意外错误：%v", err)
	}
	if !strings.HasPrefix(got, "17") || len(got) != 13 {
		t.Errorf("结果格式异常：%q", got)
	}
}
