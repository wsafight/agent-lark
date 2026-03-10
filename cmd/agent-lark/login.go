package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/auth"
)

func newLoginCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "使用内置公共应用登录（快速）",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Root().PersistentFlags().GetString("profile")
			config, _ := cmd.Root().PersistentFlags().GetString("config")

			if !auth.PublicAppAvailable() {
				fmt.Println("当前构建未内置公共应用凭据。")
				fmt.Println("请改用: agent-lark setup")
				return nil
			}

			appID, appSecret := auth.PublicAppCredentials()

			// Create a minimal config with the public app credentials
			cfg := &auth.Config{
				AppID:            appID,
				AppSecret:        appSecret,
				DefaultTokenMode: "auto",
			}

			if err := auth.Save(cfg, config, profile); err != nil {
				return fmt.Errorf("保存配置失败：%w", err)
			}

			fmt.Println("✓ 公共应用凭据已配置")

			// Determine project root for binding
			cwd, _ := os.Getwd()
			projectRoot := detectProjectRoot(cwd)
			effectiveProfile := resolveEffectiveProfile(profile, projectRoot)

			// OAuth login
			if err := auth.OAuthLogin(appID, appSecret, "docx:readonly drive:readonly", config, effectiveProfile, 9999); err != nil {
				return fmt.Errorf("OAuth 登录失败：%w", err)
			}

			_ = saveProjectBinding(projectRoot, effectiveProfile)
			fmt.Println("✓ 登录成功")
			return nil
		},
	}
}

func newSetupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "交互式配置向导（自建应用）",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Root().PersistentFlags().GetString("profile")
			config, _ := cmd.Root().PersistentFlags().GetString("config")

			reader := bufio.NewReader(os.Stdin)

			// Step 1: Choose platform
			fmt.Println("=== agent-lark 配置向导 ===")
			fmt.Println("请选择平台:")
			fmt.Println("  1. 飞书（feishu.cn）  - 开发者控制台: https://open.feishu.cn/app")
			fmt.Println("  2. Lark（larksuite.com）- 开发者控制台: https://open.larksuite.com/app")
			fmt.Print("输入选项 [1/2，默认 1]: ")
			choice, _ := reader.ReadString('\n')
			choice = strings.TrimSpace(choice)

			domain := ""
			if choice == "2" {
				domain = "open.larksuite.com"
				fmt.Println("已选择 Lark（larksuite.com）")
			} else {
				fmt.Println("已选择飞书（feishu.cn）")
			}

			// Step 2: App ID and Secret
			fmt.Print("请输入 App ID: ")
			appID, _ := reader.ReadString('\n')
			appID = strings.TrimSpace(appID)
			if appID == "" {
				return fmt.Errorf("App ID 不能为空")
			}

			fmt.Print("请输入 App Secret: ")
			appSecret, _ := reader.ReadString('\n')
			appSecret = strings.TrimSpace(appSecret)
			if appSecret == "" {
				return fmt.Errorf("App Secret 不能为空")
			}

			// Step 3: Validate credentials
			fmt.Print("正在验证凭据...")
			if err := auth.ValidateAppCredentials(appID, appSecret, domain); err != nil {
				fmt.Println(" 失败")
				return fmt.Errorf("凭据验证失败：%w", err)
			}
			fmt.Println(" ✓")

			// Step 4: Save config
			cfg := &auth.Config{
				AppID:            appID,
				AppSecret:        appSecret,
				Domain:           domain,
				DefaultTokenMode: "auto",
			}

			if err := auth.Save(cfg, config, profile); err != nil {
				return fmt.Errorf("保存配置失败：%w", err)
			}
			fmt.Println("✓ 应用凭据已保存")

			// Determine project root for binding
			cwd, _ := os.Getwd()
			projectRoot := detectProjectRoot(cwd)
			effectiveProfile := resolveEffectiveProfile(profile, projectRoot)

			// Step 5: Ask about user OAuth
			fmt.Print("是否授权用户身份（用于操作个人文档）？[y/N]: ")
			doOAuth, _ := reader.ReadString('\n')
			doOAuth = strings.TrimSpace(strings.ToLower(doOAuth))

			if doOAuth == "y" {
				if err := auth.OAuthLogin(appID, appSecret, "docx:readonly drive:readonly", config, effectiveProfile, 9999); err != nil {
					fmt.Printf("⚠ OAuth 授权失败：%s\n  可稍后运行 'agent-lark auth oauth' 重试\n", err.Error())
				}
			}

			_ = saveProjectBinding(projectRoot, effectiveProfile)
			fmt.Printf("✓ 配置完成！profile: %s\n", effectiveProfile)
			return nil
		},
	}
}
