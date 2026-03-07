package template

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var varRe = regexp.MustCompile(`\{\{(\w+)\}\}`)

// BuiltinVars 返回内置变量的当前值。
func BuiltinVars(docTitle, authorName string) map[string]string {
	now := time.Now()
	_, week := now.ISOWeek()
	return map[string]string{
		"date":     now.Format("2006-01-02"),
		"datetime": now.Format("2006-01-02 15:04"),
		"author":   authorName,
		"title":    docTitle,
		"week":     fmt.Sprintf("W%02d", week),
	}
}

// ExtractVars 提取模板中所有 {{变量名}}，区分内置和自定义。
func ExtractVars(content string) []string {
	seen := map[string]bool{}
	var vars []string
	for _, m := range varRe.FindAllStringSubmatch(content, -1) {
		name := m[1]
		if !seen[name] {
			seen[name] = true
			vars = append(vars, name)
		}
	}
	return vars
}

// Render 将模板内容中的 {{变量}} 替换为实际值。
func Render(content string, vars map[string]string) string {
	return varRe.ReplaceAllStringFunc(content, func(match string) string {
		name := strings.Trim(match, "{}")
		if v, ok := vars[name]; ok {
			return v
		}
		return match // 未提供值时保留原样
	})
}
