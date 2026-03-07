package template

import (
	"regexp"
	"testing"
	"time"
)

func TestBuiltinVars(t *testing.T) {
	got := BuiltinVars("季度报告", "Alice")

	if got["title"] != "季度报告" {
		t.Errorf("title: got %q, want %q", got["title"], "季度报告")
	}
	if got["author"] != "Alice" {
		t.Errorf("author: got %q, want %q", got["author"], "Alice")
	}

	if _, err := time.Parse("2006-01-02", got["date"]); err != nil {
		t.Errorf("date 格式错误: %q (%v)", got["date"], err)
	}
	if _, err := time.Parse("2006-01-02 15:04", got["datetime"]); err != nil {
		t.Errorf("datetime 格式错误: %q (%v)", got["datetime"], err)
	}
	if ok, _ := regexp.MatchString(`^W\d{2}$`, got["week"]); !ok {
		t.Errorf("week 格式错误: got %q, want like W09", got["week"])
	}
}

func TestExtractVars(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "提取并按出现顺序去重",
			content: "Hello {{name}}, today is {{date}}. {{name}} again.",
			want:    []string{"name", "date"},
		},
		{
			name:    "没有变量时为空",
			content: "plain text only",
			want:    nil,
		},
		{
			name:    "非单词字符变量名不匹配",
			content: "bad {{user-name}} ignored, good {{username}} kept",
			want:    []string{"username"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ExtractVars(tc.content)
			if len(got) != len(tc.want) {
				t.Fatalf("len: got %d, want %d", len(got), len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("item %d: got %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestRender(t *testing.T) {
	tests := []struct {
		name    string
		content string
		vars    map[string]string
		want    string
	}{
		{
			name:    "正常替换多个变量",
			content: "Hi {{name}}, date={{date}}",
			vars: map[string]string{
				"name": "Bob",
				"date": "2026-03-07",
			},
			want: "Hi Bob, date=2026-03-07",
		},
		{
			name:    "缺失变量保留原样",
			content: "Hi {{name}}, from {{city}}",
			vars: map[string]string{
				"name": "Bob",
			},
			want: "Hi Bob, from {{city}}",
		},
		{
			name:    "空 vars 不替换",
			content: "{{a}} {{b}}",
			vars:    map[string]string{},
			want:    "{{a}} {{b}}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Render(tc.content, tc.vars)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
