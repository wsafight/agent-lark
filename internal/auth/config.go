package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	AppID            string       `json:"app_id"`
	AppSecret        string       `json:"app_secret"`
	Domain           string       `json:"domain,omitempty"`
	DefaultTokenMode string       `json:"default_token_mode"` // auto|tenant|user
	UserSession      *UserSession `json:"user_session,omitempty"`
}

type UserSession struct {
	OpenID           string `json:"open_id"`
	Name             string `json:"name"`
	UserAccessToken  string `json:"user_access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresAt        string `json:"expires_at"`
	RefreshExpiresAt string `json:"refresh_expires_at,omitempty"`
}

func HomeDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".agent-lark")
}

func ProfilesDir() string {
	return filepath.Join(HomeDir(), "profiles")
}

func TemplatesDir() string {
	return filepath.Join(HomeDir(), "templates")
}

func ProfileConfigPath(profile string) string {
	if profile == "" {
		profile = "default"
	}
	return filepath.Join(ProfilesDir(), profile+".json")
}

func ResolveConfigPath(explicitConfig, profile string) string {
	if explicitConfig != "" {
		return explicitConfig
	}
	return ProfileConfigPath(ResolveEffectiveProfile(profile))
}

func Load(explicitConfig, profile string) (*Config, string, error) {
	path := ResolveConfigPath(explicitConfig, profile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, path, fmt.Errorf("AUTH_REQUIRED：尚未配置凭据，请运行 'agent-lark login'（或 'agent-lark setup'）")
		}
		return nil, path, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		backupPath := path + ".broken"
		_ = os.Rename(path, backupPath)
		return nil, path, fmt.Errorf(
			"配置文件已损坏，已备份至 %s\n请运行 'agent-lark login' 或 'agent-lark setup' 重新配置。\n解析错误：%v",
			backupPath, err,
		)
	}
	if key, err := loadOrCreateMasterKey(); err == nil {
		cfg.AppSecret, _ = decryptField(key, cfg.AppSecret)
		if cfg.UserSession != nil {
			cfg.UserSession.UserAccessToken, _ = decryptField(key, cfg.UserSession.UserAccessToken)
			cfg.UserSession.RefreshToken, _ = decryptField(key, cfg.UserSession.RefreshToken)
		}
	}

	return &cfg, path, nil
}

func Save(cfg *Config, explicitConfig, profile string) error {
	toSave := *cfg

	key, err := loadOrCreateMasterKey()
	if err != nil {
		fmt.Fprintln(os.Stderr, "⚠ 无法初始化加密密钥，凭据以明文存储："+err.Error())
	} else {
		toSave.AppSecret, _ = encryptField(key, cfg.AppSecret)
		if cfg.UserSession != nil {
			sess := *cfg.UserSession
			sess.UserAccessToken, _ = encryptField(key, cfg.UserSession.UserAccessToken)
			sess.RefreshToken, _ = encryptField(key, cfg.UserSession.RefreshToken)
			toSave.UserSession = &sess
		}
	}

	path := ResolveConfigPath(explicitConfig, profile)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(toSave, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func ListProfiles() ([]string, error) {
	entries, err := os.ReadDir(ProfilesDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var profiles []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			profiles = append(profiles, e.Name()[:len(e.Name())-5])
		}
	}
	return profiles, nil
}
