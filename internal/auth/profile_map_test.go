package auth

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestDetectProjectRoot(t *testing.T) {
	t.Run("含 .git 标记的目录", func(t *testing.T) {
		tmp := t.TempDir()
		// 创建嵌套目录 tmp/a/b/c，在 tmp/a 放 .git
		projRoot := filepath.Join(tmp, "a")
		nested := filepath.Join(projRoot, "b", "c")
		if err := os.MkdirAll(nested, 0755); err != nil {
			t.Fatalf("创建目录失败：%v", err)
		}
		if err := os.Mkdir(filepath.Join(projRoot, ".git"), 0755); err != nil {
			t.Fatalf("创建 .git 失败：%v", err)
		}

		got := DetectProjectRoot(nested)
		if got != projRoot {
			t.Errorf("DetectProjectRoot = %q, want %q", got, projRoot)
		}
	})

	t.Run("无标记回退到 cwd", func(t *testing.T) {
		tmp := t.TempDir()
		got := DetectProjectRoot(tmp)
		// tmp 本身无标记，向上也找不到（到根目录为止），应回退到 cwd
		if got != tmp {
			t.Errorf("DetectProjectRoot = %q, want %q (回退到 cwd)", got, tmp)
		}
	})
}

func TestProjectHashProfile(t *testing.T) {
	t.Run("确定性", func(t *testing.T) {
		a := ProjectHashProfile("/some/path")
		b := ProjectHashProfile("/some/path")
		if a != b {
			t.Errorf("same input produced different results: %q vs %q", a, b)
		}
	})

	t.Run("不同路径产生不同 hash", func(t *testing.T) {
		a := ProjectHashProfile("/path/a")
		b := ProjectHashProfile("/path/b")
		if a == b {
			t.Errorf("different inputs produced same result: %q", a)
		}
	})

	t.Run("格式校验 project-[0-9a-f]{16}", func(t *testing.T) {
		got := ProjectHashProfile("/any/project")
		re := regexp.MustCompile(`^project-[0-9a-f]{16}$`)
		if !re.MatchString(got) {
			t.Errorf("ProjectHashProfile = %q, does not match project-[0-9a-f]{16}", got)
		}
	})
}

func TestResolveEffectiveProfile(t *testing.T) {
	// 清理环境变量避免干扰
	origEnv := os.Getenv("AGENT_LARK_PROFILE")
	os.Unsetenv("AGENT_LARK_PROFILE")
	defer os.Setenv("AGENT_LARK_PROFILE", origEnv)

	t.Run("显式 profile 优先", func(t *testing.T) {
		got := ResolveEffectiveProfile("staging")
		if got != "staging" {
			t.Errorf("got %q, want %q", got, "staging")
		}
	})

	t.Run("空 profile 返回 project hash", func(t *testing.T) {
		got := ResolveEffectiveProfile("")
		re := regexp.MustCompile(`^project-[0-9a-f]{16}$`)
		if !re.MatchString(got) {
			t.Errorf("got %q, want project-[0-9a-f]{16} format", got)
		}
	})

	t.Run("环境变量优先于默认", func(t *testing.T) {
		os.Setenv("AGENT_LARK_PROFILE", "from-env")
		defer os.Unsetenv("AGENT_LARK_PROFILE")

		got := ResolveEffectiveProfile("")
		if got != "from-env" {
			t.Errorf("got %q, want %q", got, "from-env")
		}
	})
}
