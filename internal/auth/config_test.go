package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// overrideHomeDir 临时替换 HomeDir 返回的路径，用于隔离测试。
// 返回的 cleanup 函数负责还原。
func overrideHomeDir(t *testing.T) (tmpHome string, cleanup func()) {
	t.Helper()
	tmpHome = t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	return tmpHome, func() { os.Setenv("HOME", origHome) }
}

func TestSaveAndLoad(t *testing.T) {
	_, cleanup := overrideHomeDir(t)
	defer cleanup()

	cfg := &Config{
		AppID:            "cli_test_id",
		AppSecret:        "test_secret_value",
		Domain:           "open.feishu.cn",
		DefaultTokenMode: "tenant",
	}

	if err := Save(cfg, "", "test-profile"); err != nil {
		t.Fatalf("保存失败：%v", err)
	}

	loaded, _, err := Load("", "test-profile")
	if err != nil {
		t.Fatalf("加载失败：%v", err)
	}
	if loaded.AppID != cfg.AppID {
		t.Errorf("AppID：got %q, want %q", loaded.AppID, cfg.AppID)
	}
	if loaded.AppSecret != cfg.AppSecret {
		t.Errorf("AppSecret：got %q, want %q", loaded.AppSecret, cfg.AppSecret)
	}
	if loaded.Domain != cfg.Domain {
		t.Errorf("Domain：got %q, want %q", loaded.Domain, cfg.Domain)
	}
	if loaded.DefaultTokenMode != cfg.DefaultTokenMode {
		t.Errorf("DefaultTokenMode：got %q, want %q", loaded.DefaultTokenMode, cfg.DefaultTokenMode)
	}
}

func TestSaveAndLoadWithUserSession(t *testing.T) {
	_, cleanup := overrideHomeDir(t)
	defer cleanup()

	cfg := &Config{
		AppID:            "cli_sess_id",
		AppSecret:        "sess_secret",
		DefaultTokenMode: "auto",
		UserSession: &UserSession{
			OpenID:          "ou_abc123",
			Name:            "测试用户",
			UserAccessToken: "u-user_token_value",
			RefreshToken:    "ur-refresh_token_value",
			ExpiresAt:       "2026-12-31T23:59:59Z",
		},
	}

	if err := Save(cfg, "", "sess-profile"); err != nil {
		t.Fatalf("保存失败：%v", err)
	}

	loaded, _, err := Load("", "sess-profile")
	if err != nil {
		t.Fatalf("加载失败：%v", err)
	}
	if loaded.UserSession == nil {
		t.Fatal("UserSession 应不为 nil")
	}
	if loaded.UserSession.OpenID != cfg.UserSession.OpenID {
		t.Errorf("OpenID：got %q, want %q", loaded.UserSession.OpenID, cfg.UserSession.OpenID)
	}
	if loaded.UserSession.UserAccessToken != cfg.UserSession.UserAccessToken {
		t.Errorf("UserAccessToken 往返不一致")
	}
	if loaded.UserSession.RefreshToken != cfg.UserSession.RefreshToken {
		t.Errorf("RefreshToken 往返不一致")
	}
}

func TestLoadNotExist(t *testing.T) {
	_, cleanup := overrideHomeDir(t)
	defer cleanup()

	_, _, err := Load("", "nonexistent")
	if err == nil {
		t.Fatal("期望错误，但未返回")
	}
	if !strings.Contains(err.Error(), "AUTH_REQUIRED") {
		t.Errorf("错误信息应含 AUTH_REQUIRED：%v", err)
	}
}

func TestLoadCorruptJSON(t *testing.T) {
	_, cleanup := overrideHomeDir(t)
	defer cleanup()

	path := ProfileConfigPath("corrupt")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatalf("创建目录失败：%v", err)
	}
	if err := os.WriteFile(path, []byte("{invalid json"), 0600); err != nil {
		t.Fatalf("写入失败：%v", err)
	}

	_, _, err := Load("", "corrupt")
	if err == nil {
		t.Fatal("期望解析错误，但未返回")
	}
	if !strings.Contains(err.Error(), "损坏") {
		t.Errorf("错误信息应含 '损坏'：%v", err)
	}
	// 验证备份文件已生成
	if _, statErr := os.Stat(path + ".broken"); statErr != nil {
		t.Errorf("备份文件未生成：%v", statErr)
	}
}

func TestProfileConfigPath(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		wantEnd string
	}{
		{"空 profile 默认 default", "", "default.json"},
		{"指定 profile", "staging", "staging.json"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ProfileConfigPath(tc.profile)
			if !strings.HasSuffix(got, tc.wantEnd) {
				t.Errorf("路径应以 %q 结尾：got %q", tc.wantEnd, got)
			}
		})
	}
}

func TestResolveConfigPath(t *testing.T) {
	explicit := "/tmp/my-config.json"
	got := ResolveConfigPath(explicit, "any-profile")
	if got != explicit {
		t.Errorf("显式 config 应优先：got %q, want %q", got, explicit)
	}
}

func TestListProfiles(t *testing.T) {
	_, cleanup := overrideHomeDir(t)
	defer cleanup()

	dir := ProfilesDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("创建目录失败：%v", err)
	}
	for _, name := range []string{"alpha.json", "beta.json", "not-json.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0600); err != nil {
			t.Fatalf("写入 %s 失败：%v", name, err)
		}
	}

	profiles, err := ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles 失败：%v", err)
	}

	want := map[string]bool{"alpha": true, "beta": true}
	if len(profiles) != len(want) {
		t.Fatalf("期望 %d 个 profile，got %d：%v", len(want), len(profiles), profiles)
	}
	for _, p := range profiles {
		if !want[p] {
			t.Errorf("意外 profile：%q", p)
		}
	}
}
