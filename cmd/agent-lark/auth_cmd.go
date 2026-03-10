package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/auth"
	"github.com/wsafight/agent-lark/internal/output"
)

// NewAuthCommand returns the auth subcommand group.
func NewAuthCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "auth", Short: "认证管理"}
	cmd.AddCommand(
		newAuthLoginCommand(),
		newAuthOAuthCommand(),
		newAuthStatusCommand(),
		newAuthProfileCommand(),
		newAuthSetModeCommand(),
		newAuthLogoutCommand(),
	)
	return cmd
}

func newAuthLoginCommand() *cobra.Command {
	var appID string
	var appSecret string
	var domain string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "配置应用凭据",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Root().PersistentFlags().GetString("profile")
			config, _ := cmd.Root().PersistentFlags().GetString("config")

			reader := bufio.NewReader(os.Stdin)

			// Interactive prompts for missing values
			if appID == "" {
				fmt.Print("请输入 App ID: ")
				line, _ := reader.ReadString('\n')
				appID = strings.TrimSpace(line)
			}
			if appSecret == "" {
				fmt.Print("请输入 App Secret: ")
				line, _ := reader.ReadString('\n')
				appSecret = strings.TrimSpace(line)
			}

			if appID == "" || appSecret == "" {
				return fmt.Errorf("App ID 和 App Secret 不能为空")
			}

			// Validate
			fmt.Print("正在验证凭据...")
			if err := auth.ValidateAppCredentials(appID, appSecret, domain); err != nil {
				fmt.Println(" 失败")
				return fmt.Errorf("凭据验证失败：%w", err)
			}
			fmt.Println(" ✓")

			cfg := &auth.Config{
				AppID:            appID,
				AppSecret:        appSecret,
				Domain:           domain,
				DefaultTokenMode: "auto",
			}

			if err := auth.Save(cfg, config, profile); err != nil {
				return fmt.Errorf("保存配置失败：%w", err)
			}

			cwd, _ := os.Getwd()
			projectRoot := detectProjectRoot(cwd)
			effectiveProfile := resolveEffectiveProfile(profile, projectRoot)
			_ = saveProjectBinding(projectRoot, effectiveProfile)

			fmt.Printf("✓ 凭据已保存（profile: %s）\n", effectiveProfile)
			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app-id", "", "飞书应用 App ID")
	cmd.Flags().StringVar(&appSecret, "app-secret", "", "飞书应用 App Secret")
	cmd.Flags().StringVar(&domain, "domain", "", "API 域名（默认 open.feishu.cn）")
	return cmd
}

func newAuthOAuthCommand() *cobra.Command {
	var scope string
	var port int

	cmd := &cobra.Command{
		Use:   "oauth",
		Short: "追加用户授权（OAuth）",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Root().PersistentFlags().GetString("profile")
			config, _ := cmd.Root().PersistentFlags().GetString("config")

			cfg, _, err := auth.Load(config, profile)
			if err != nil {
				return err
			}

			if cfg.AppID == "" || cfg.AppSecret == "" {
				return fmt.Errorf("AUTH_REQUIRED：请先运行 'agent-lark setup' 配置应用凭据")
			}

			if err := auth.OAuthLogin(cfg.AppID, cfg.AppSecret, scope, config, profile, port); err != nil {
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
			profile, _ := cmd.Root().PersistentFlags().GetString("profile")
			config, _ := cmd.Root().PersistentFlags().GetString("config")
			format, _ := cmd.Root().PersistentFlags().GetString("format")
			agent, _ := cmd.Root().PersistentFlags().GetBool("agent")
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)

			cwd, _ := os.Getwd()
			projectRoot := detectProjectRoot(cwd)
			effectiveProfile := resolveEffectiveProfile(profile, projectRoot)

			cfg, cfgPath, err := auth.Load(config, effectiveProfile)
			if err != nil {
				return err
			}

			type tokenStatus struct {
				Name      string `json:"name,omitempty"`
				ExpiresAt string `json:"expires_at,omitempty"`
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
				valid := time.Now().Before(expiresAt)
				result.User = &tokenStatus{
					Name:      cfg.UserSession.Name,
					ExpiresAt: cfg.UserSession.ExpiresAt,
					Valid:     valid,
				}
			}

			if format == "json" {
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
				status := "有效"
				if !result.User.Valid {
					status = "已过期"
				}
				fmt.Printf("用户 Token: %s（%s，%s）\n", result.User.Name, status, result.User.ExpiresAt)
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
			format, _ := cmd.Root().PersistentFlags().GetString("format")
			agent, _ := cmd.Root().PersistentFlags().GetBool("agent")
			if agent {
				output.GlobalAgent = true
				format = "json"
			}
			format = output.FormatFromCmd(format)

			profiles, err := auth.ListProfiles()
			if err != nil {
				return fmt.Errorf("列出 profiles 失败：%w", err)
			}

			if format == "json" {
				return output.PrintJSON(os.Stdout, profiles)
			}

			if len(profiles) == 0 {
				fmt.Println("（无 profiles）")
				return nil
			}
			for _, p := range profiles {
				fmt.Println(p)
			}
			return nil
		},
	}

	useCmd := &cobra.Command{
		Use:   "use <name>",
		Short: "切换当前项目的 profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			cwd, _ := os.Getwd()
			projectRoot := detectProjectRoot(cwd)

			if err := saveProjectBinding(projectRoot, profileName); err != nil {
				return fmt.Errorf("保存 profile 绑定失败：%w", err)
			}

			fmt.Printf("✓ 项目 %s 已绑定到 profile: %s\n", projectRoot, profileName)
			return nil
		},
	}

	cmd.AddCommand(listCmd, useCmd)
	return cmd
}

func newAuthSetModeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-mode <auto|tenant|user>",
		Short: "修改默认 Token 模式",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Root().PersistentFlags().GetString("profile")
			config, _ := cmd.Root().PersistentFlags().GetString("config")

			mode := args[0]
			if mode != "auto" && mode != "tenant" && mode != "user" {
				return fmt.Errorf("无效的 Token 模式：%s（支持 auto|tenant|user）", mode)
			}

			cfg, _, err := auth.Load(config, profile)
			if err != nil {
				return err
			}

			cfg.DefaultTokenMode = mode
			if err := auth.Save(cfg, config, profile); err != nil {
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
			profile, _ := cmd.Root().PersistentFlags().GetString("profile")
			config, _ := cmd.Root().PersistentFlags().GetString("config")

			if all {
				// Full reset: remove the config file
				cfgPath := auth.ResolveConfigPath(config, profile)
				if err := os.Remove(cfgPath); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("删除配置失败：%w", err)
				}
				fmt.Println("✓ 已清除所有认证信息")
				return nil
			}

			if userOnly {
				cfg, _, err := auth.Load(config, profile)
				if err != nil {
					return err
				}
				cfg.UserSession = nil
				if err := auth.Save(cfg, config, profile); err != nil {
					return fmt.Errorf("保存配置失败：%w", err)
				}
				fmt.Println("✓ 已清除用户 Token")
				return nil
			}

			// Default: clear user token
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("清除用户 Token（--user）还是全部重置（--all）？输入 'all' 全部重置，其他清除用户 Token: ")
			choice, _ := reader.ReadString('\n')
			choice = strings.TrimSpace(strings.ToLower(choice))

			cfg, _, err := auth.Load(config, profile)
			if err != nil {
				return err
			}

			if choice == "all" {
				cfgPath := auth.ResolveConfigPath(config, profile)
				if err := os.Remove(cfgPath); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("删除配置失败：%w", err)
				}
				fmt.Println("✓ 已清除所有认证信息")
			} else {
				cfg.UserSession = nil
				if err := auth.Save(cfg, config, profile); err != nil {
					return fmt.Errorf("保存配置失败：%w", err)
				}
				fmt.Println("✓ 已清除用户 Token")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&userOnly, "user", false, "仅清除用户 Token")
	cmd.Flags().BoolVar(&all, "all", false, "全部重置（包括应用凭据）")
	return cmd
}
