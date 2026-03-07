package template

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wangshian/agent-lark/internal/auth"
)

type Template struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Content     string    `json:"content"` // Markdown 内容
	Source      string    `json:"source,omitempty"` // "local" 或文档 URL
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func TemplateDir() string {
	return auth.TemplatesDir()
}

func templatePath(name string) string {
	return filepath.Join(TemplateDir(), name+".json")
}

func Save(t *Template, force bool) error {
	path := templatePath(t.Name)
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("TEMPLATE_EXISTS：模板 %q 已存在，使用 --force 覆盖", t.Name)
		}
	}
	if err := os.MkdirAll(TemplateDir(), 0700); err != nil {
		return err
	}
	data, _ := json.MarshalIndent(t, "", "  ")
	return os.WriteFile(path, data, 0600)
}

func Load(name string) (*Template, error) {
	data, err := os.ReadFile(templatePath(name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("TEMPLATE_NOT_FOUND：模板 %q 不存在，运行 'agent-lark template list' 查看可用模板", name)
		}
		return nil, err
	}
	var t Template
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func Delete(name string) error {
	path := templatePath(name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("TEMPLATE_NOT_FOUND：模板 %q 不存在", name)
	}
	return os.Remove(path)
}

func ListAll() ([]*Template, error) {
	entries, err := os.ReadDir(TemplateDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var templates []*Template
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			name := e.Name()[:len(e.Name())-5]
			t, err := Load(name)
			if err == nil {
				templates = append(templates, t)
			}
		}
	}
	return templates, nil
}
