package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/auth"
	"github.com/wsafight/agent-lark/internal/cmdutil"
	"github.com/wsafight/agent-lark/internal/output"
)

// NewAuthCommand returns the auth subcommand group.
func NewAuthCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "auth", Short: "认证管理"}
	cmd.AddCommand(
		newAuthPublicLoginCommand(),
		newAuthOAuthCommand(),
		newAuthStatusCommand(),
		newAuthProfileCommand(),
		newAuthSetModeCommand(),
		newAuthLogoutCommand(),
	)
	return cmd
}

// newAuthPublicLoginCommand 使用内置公共应用快速登录（仅特殊构建可用）。
func newAuthPublicLoginCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "使用内置公共应用登录（快速，仅特定构建可用）",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.GetGlobalFlags(cmd)

			if !auth.PublicAppAvailable() {
				fmt.Println("当前构建未内置公共应用凭据。")
				fmt.Println("请改用: agent-lark setup")
				return nil
			}

			appID, appSecret := auth.PublicAppCredentials()

			cfg := &auth.Config{
				AppID:            appID,
				AppSecret:        appSecret,
				DefaultTokenMode: "auto",
			}
			if err := auth.Save(cfg, g.Config, g.Profile); err != nil {
				return fmt.Errorf("保存配置失败：%w", err)
			}
			fmt.Println("✓ 公共应用凭据已配置")

			effectiveProfile := auth.ResolveEffectiveProfile(g.Profile)

			if err := auth.OAuthLogin(appID, appSecret, oauthScope, g.Config, effectiveProfile, oauthCallbackPort); err != nil {
				return fmt.Errorf("OAuth 登录失败：%w", err)
			}

			fmt.Println("✓ 登录成功")
			return nil
		},
	}
}


func newAuthOAuthCommand() *cobra.Command {
	var scope string
	var port int

	cmd := &cobra.Command{
		Use:   "oauth",
		Short: "追加用户授权（OAuth）",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.GetGlobalFlags(cmd)

			cfg, _, err := auth.Load(g.Config, g.Profile)
			if err != nil {
				return err
			}

			if cfg.AppID == "" || cfg.AppSecret == "" {
				return fmt.Errorf("AUTH_REQUIRED：请先运行 'agent-lark setup' 配置应用凭据")
			}

			if err := auth.OAuthLogin(cfg.AppID, cfg.AppSecret, scope, g.Config, g.Profile, port); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "docx:readonly drive:readonly", "OAuth 授权范围")
	cmd.Flags().IntVar(&port, "port", 9999, "本地回调端口")
	return cmd
}

func newAuthStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "显示当前认证状态",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.ResolveGlobalFlags(cmd)
			effectiveProfile := auth.ResolveEffectiveProfile(g.Profile)

			cfg, cfgPath, err := auth.Load(g.Config, effectiveProfile)
			if err != nil {
				return err
			}

			type tokenStatus struct {
				Name      string `json:"name,omitempty"`
				ExpiresAt string `json:"expires_at,omitempty"`
				Remaining string `json:"remaining,omitempty"`
				Valid     bool   `json:"valid"`
			}

			type statusResult struct {
				Profile    string       `json:"profile"`
				ConfigPath string       `json:"config_path"`
				AppID      string       `json:"app_id"`
				Domain     string       `json:"domain,omitempty"`
				TokenMode  string       `json:"token_mode"`
				User       *tokenStatus `json:"user,omitempty"`
			}

			result := statusResult{
				Profile:    effectiveProfile,
				ConfigPath: cfgPath,
				AppID:      cfg.AppID,
				Domain:     cfg.Domain,
				TokenMode:  cfg.DefaultTokenMode,
			}

			if cfg.UserSession != nil && cfg.UserSession.UserAccessToken != "" {
				expiresAt, _ := time.Parse(time.RFC3339, cfg.UserSession.ExpiresAt)
				now := time.Now()
				valid := now.Before(expiresAt)
				ts := &tokenStatus{
					Name:      cfg.UserSession.Name,
					ExpiresAt: cfg.UserSession.ExpiresAt,
					Valid:     valid,
				}
				if valid {
					ts.Remaining = expiresAt.Sub(now).Round(time.Minute).String()
				}
				result.User = ts
			}

			if g.Format == "json" {
				return output.PrintJSON(os.Stdout, result)
			}

			fmt.Printf("Profile:    %s\n", result.Profile)
			fmt.Printf("配置文件:   %s\n", result.ConfigPath)
			fmt.Printf("App ID:     %s\n", result.AppID)
			if result.Domain != "" {
				fmt.Printf("Domain:     %s\n", result.Domain)
			}
			fmt.Printf("Token 模式: %s\n", result.TokenMode)
			if result.User != nil {
				if result.User.Valid {
					fmt.Printf("用户 Token: %s（有效，剩余 %s）\n", result.User.Name, result.User.Remaining)
				} else {
					fmt.Printf("用户 Token: %s（已过期，运行 'agent-lark auth oauth' 刷新）\n", result.User.Name)
				}
			} else {
				fmt.Println("用户 Token: 未配置（运行 'agent-lark auth oauth' 授权）")
			}
			return nil
		},
	}
}

func newAuthProfileCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "profile", Short: "Profile 管理"}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有 profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.ResolveGlobalFlags(cmd)

			profiles, err := auth.ListProfiles()
			if err != nil {
				return fmt.Errorf("列出 profiles 失败：%w", err)
			}

			if g.Format == "json" {
				return output.PrintJSON(os.Stdout, profiles)
			}

			if len(profiles) == 0 {
				fmt.Println("（无 profiles）")
				return nil
			}
			active := auth.ResolveEffectiveProfile("")
			for _, p := range profiles {
				if p == active {
					fmt.Printf("* %s  (当前)\n", p)
				} else {
					fmt.Printf("  %s\n", p)
				}
			}
			return nil
		},
	}

	cmd.AddCommand(listCmd)
	return cmd
}

func newAuthSetModeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-mode <auto|tenant|user>",
		Short: "修改默认 Token 模式",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.GetGlobalFlags(cmd)

			mode := args[0]
			if mode != "auto" && mode != "tenant" && mode != "user" {
				return fmt.Errorf("无效的 Token 模式：%s（支持 auto|tenant|user）", mode)
			}

			cfg, _, err := auth.Load(g.Config, g.Profile)
			if err != nil {
				return err
			}

			cfg.DefaultTokenMode = mode
			if err := auth.Save(cfg, g.Config, g.Profile); err != nil {
				return fmt.Errorf("保存配置失败：%w", err)
			}

			fmt.Printf("✓ Token 模式已设置为: %s\n", mode)
			return nil
		},
	}
}

func newAuthLogoutCommand() *cobra.Command {
	var userOnly bool
	var all bool

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "退出登录",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.GetGlobalFlags(cmd)

			if all {
				cfgPath := auth.ResolveConfigPath(g.Config, g.Profile)
				if err := os.Remove(cfgPath); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("删除配置失败：%w", err)
				}
				fmt.Println("✓ 已清除所有认证信息")
				return nil
			}

			// Default (no flags) and --user: clear user token only
			cfg, _, err := auth.Load(g.Config, g.Profile)
			if err != nil {
				return err
			}
			cfg.UserSession = nil
			if err := auth.Save(cfg, g.Config, g.Profile); err != nil {
				return fmt.Errorf("保存配置失败：%w", err)
			}
			fmt.Println("✓ 已清除用户 Token（使用 --all 可同时清除应用凭据）")
			return nil
		},
	}

	cmd.Flags().BoolVar(&userOnly, "user", false, "仅清除用户 Token（默认行为）")
	cmd.Flags().BoolVar(&all, "all", false, "全部重置（包括应用凭据）")
	return cmd
}
