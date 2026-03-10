package client

import (
	"testing"

	"github.com/wsafight/agent-lark/internal/auth"
)

func TestRequestOptions(t *testing.T) {
	t.Run("user 模式返回 UserAccessToken 选项", func(t *testing.T) {
		r := &Result{
			Mode:      "user",
			UserToken: "u-test-token",
		}
		opts := r.RequestOptions()
		if len(opts) != 1 {
			t.Fatalf("期望 1 个选项，got %d", len(opts))
		}
	})

	t.Run("tenant 模式返回 nil", func(t *testing.T) {
		r := &Result{
			Mode: "tenant",
		}
		opts := r.RequestOptions()
		if opts != nil {
			t.Errorf("tenant 模式应返回 nil，got %v", opts)
		}
	})

	t.Run("user 模式但 token 为空返回 nil", func(t *testing.T) {
		r := &Result{
			Mode:      "user",
			UserToken: "",
		}
		opts := r.RequestOptions()
		if opts != nil {
			t.Errorf("user 模式 token 为空应返回 nil，got %v", opts)
		}
	})
}

func TestBuildResultTokenModeResolution(t *testing.T) {
	baseCfg := &auth.Config{
		AppID:            "cli_test",
		AppSecret:        "secret",
		DefaultTokenMode: "tenant",
	}

	tests := []struct {
		name     string
		cfg      *auth.Config
		opts     Options
		wantMode string
	}{
		{
			name:     "tenant 模式直通",
			cfg:      baseCfg,
			opts:     Options{TokenMode: "tenant"},
			wantMode: "tenant",
		},
		{
			name: "auto 模式无 UserSession 回退 tenant",
			cfg: &auth.Config{
				AppID:            "cli_test",
				AppSecret:        "secret",
				DefaultTokenMode: "auto",
			},
			opts:     Options{TokenMode: "auto"},
			wantMode: "tenant",
		},
		{
			name:     "空 mode 继承 config 的 tenant",
			cfg:      baseCfg,
			opts:     Options{TokenMode: ""},
			wantMode: "tenant",
		},
		{
			name: "空 mode + config 也为空 → auto → tenant (无 session)",
			cfg: &auth.Config{
				AppID:     "cli_test",
				AppSecret: "secret",
			},
			opts:     Options{TokenMode: ""},
			wantMode: "tenant",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := buildResult(tc.cfg, tc.opts)
			if err != nil {
				t.Fatalf("buildResult 失败：%v", err)
			}
			if result.Mode != tc.wantMode {
				t.Errorf("Mode：got %q, want %q", result.Mode, tc.wantMode)
			}
		})
	}
}
