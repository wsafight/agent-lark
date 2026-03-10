package template

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wsafight/agent-lark/internal/client"
)

func TestResolveTemplateSourceInline(t *testing.T) {
	gotContent, gotSource, err := resolveTemplateSource(context.Background(), "", "", "hello", client.Options{})
	if err != nil {
		t.Fatalf("resolveTemplateSource error: %v", err)
	}
	if gotContent != "hello" {
		t.Fatalf("content: got %q, want %q", gotContent, "hello")
	}
	if gotSource != "local" {
		t.Fatalf("source: got %q, want %q", gotSource, "local")
	}
}

func TestResolveTemplateSourceFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "template.md")
	if err := os.WriteFile(path, []byte("from-file"), 0600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	gotContent, gotSource, err := resolveTemplateSource(context.Background(), "", path, "", client.Options{})
	if err != nil {
		t.Fatalf("resolveTemplateSource error: %v", err)
	}
	if gotContent != "from-file" {
		t.Fatalf("content: got %q, want %q", gotContent, "from-file")
	}
	if gotSource != "local" {
		t.Fatalf("source: got %q, want %q", gotSource, "local")
	}
}

func TestResolveTemplateSourceMissingInput(t *testing.T) {
	_, _, err := resolveTemplateSource(context.Background(), "", "", "", client.Options{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "MISSING_FLAG") {
		t.Fatalf("error: got %q, want contains %q", err.Error(), "MISSING_FLAG")
	}
}

func TestApplyInvalidTargetURLBeforeClientInit(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	now := time.Now()
	err := Save(&Template{
		Name:      "demo",
		Content:   "hello",
		CreatedAt: now,
		UpdatedAt: now,
	}, true)
	if err != nil {
		t.Fatalf("save template: %v", err)
	}

	_, err = Apply(context.Background(), ApplyOptions{
		TemplateName: "demo",
		TargetURL:    "https://example.com/not-a-doc-url",
		ClientOpts: client.Options{
			TokenMode: "tenant",
		},
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "INVALID_URL") {
		t.Fatalf("error: got %q, want contains %q", err.Error(), "INVALID_URL")
	}
	if strings.Contains(err.Error(), "AUTH_REQUIRED") {
		t.Fatalf("unexpected auth path: %q", err.Error())
	}
}
