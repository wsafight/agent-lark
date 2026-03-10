package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/auth"
	"github.com/wsafight/agent-lark/internal/cmdutil"
)

const (
	oauthScope        = "docx:readonly drive:readonly"
	oauthCallbackPort = 9999
)


func runSetupWizard(config, profile string) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	// 加载已有配置（新用户时忽略错误）
	existing, _, _ := auth.Load(config, profile)

	fmt.Println("=== agent-lark 配置向导 ===")
	fmt.Println("（已有值直接回车跳过）")
	fmt.Println()

	// Step 1: 选择平台
	isLark := existing != nil && existing.Domain == "open.larksuite.com"
	fmt.Println("请选择平台:")
	if isLark {
		fmt.Println("  1. 飞书（feishu.cn）  - https://open.feishu.cn/app")
		fmt.Println("  2. Lark（larksuite.com）- https://open.larksuite.com/app  ← 当前")
	} else {
		fmt.Println("  1. 飞书（feishu.cn）  - https://open.feishu.cn/app  ← 当前")
		fmt.Println("  2. Lark（larksuite.com）- https://open.larksuite.com/app")
	}
	fmt.Print("输入选项 [1/2，回车保持]: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	domain := ""
	if existing != nil {
		domain = existing.Domain
	}
	switch choice {
	case "1":
		domain = ""
		fmt.Println("已选择飞书（feishu.cn）")
	case "2":
		domain = "open.larksuite.com"
		fmt.Println("已选择 Lark（larksuite.com）")
	default:
		if isLark {
			fmt.Println("保持 Lark（larksuite.com）")
		} else {
			fmt.Println("保持飞书（feishu.cn）")
		}
	}

	// Step 2: App ID
	currentAppID := ""
	if existing != nil {
		currentAppID = existing.AppID
	}
	if currentAppID != "" {
		fmt.Printf("请输入 App ID [当前: %s，回车跳过]: ", currentAppID)
	} else {
		fmt.Print("请输入 App ID: ")
	}
	appIDInput, _ := reader.ReadString('\n')
	appIDInput = strings.TrimSpace(appIDInput)

	appID, idChanged := appIDInput, true
	if appIDInput == "" {
		if currentAppID == "" {
			return "", fmt.Errorf("INVALID_INPUT：App ID 不能为空")
		}
		appID, idChanged = currentAppID, false
	}

	// Step 3: App Secret
	hasSecret := existing != nil && existing.AppSecret != ""
	if hasSecret {
		fmt.Print("请输入 App Secret [已设置，回车跳过]: ")
	} else {
		fmt.Print("请输入 App Secret: ")
	}
	appSecretInput, _ := reader.ReadString('\n')
	appSecretInput = strings.TrimSpace(appSecretInput)

	appSecret, secretChanged := appSecretInput, true
	if appSecretInput == "" {
		if !hasSecret {
			return "", fmt.Errorf("INVALID_INPUT：App Secret 不能为空")
		}
		appSecret, secretChanged = existing.AppSecret, false
	}

	// Step 4: 仅凭据有变更时重新验证
	if idChanged || secretChanged {
		fmt.Print("正在验证凭据...")
		if err := auth.ValidateAppCredentials(appID, appSecret, domain); err != nil {
			fmt.Println(" 失败")
			return "", fmt.Errorf("凭据验证失败：%w", err)
		}
		fmt.Println(" ✓")
	} else {
		fmt.Println("凭据未变更，跳过验证")
	}

	// Step 5: 保存配置，保留已有的 DefaultTokenMode
	tokenMode := "auto"
	if existing != nil && existing.DefaultTokenMode != "" {
		tokenMode = existing.DefaultTokenMode
	}
	cfg := &auth.Config{
		AppID:            appID,
		AppSecret:        appSecret,
		Domain:           domain,
		DefaultTokenMode: tokenMode,
	}
	if err := auth.Save(cfg, config, profile); err != nil {
		return "", fmt.Errorf("保存配置失败：%w", err)
	}
	fmt.Println("✓ 应用凭据已保存")

	effectiveProfile := auth.ResolveEffectiveProfile(profile)

	// Step 6: 用户 OAuth
	hasOAuth := existing != nil && existing.UserSession != nil && existing.UserSession.UserAccessToken != ""
	credChanged := idChanged || secretChanged
	if hasOAuth && !credChanged {
		fmt.Printf("是否重新授权用户身份？[当前: %s，y/N]: ", existing.UserSession.Name)
	} else {
		fmt.Print("是否授权用户身份（用于操作个人文档）？[y/N]: ")
	}
	doOAuth, _ := reader.ReadString('\n')
	doOAuth = strings.TrimSpace(strings.ToLower(doOAuth))

	if doOAuth == "y" {
		if err := auth.OAuthLogin(appID, appSecret, oauthScope, config, effectiveProfile, oauthCallbackPort); err != nil {
			fmt.Printf("⚠ OAuth 授权失败：%s\n  可稍后运行 'agent-lark auth oauth' 重试\n", err.Error())
		}
	} else if !hasOAuth {
		fmt.Println("  （跳过，可稍后运行 'agent-lark auth oauth' 授权）")
	}

	return effectiveProfile, nil
}

func newSetupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "交互式配置向导（自建应用）",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.GetGlobalFlags(cmd)
			effectiveProfile, err := runSetupWizard(g.Config, g.Profile)
			if err != nil {
				return err
			}
			fmt.Printf("✓ 配置完成！profile: %s\n", effectiveProfile)
			return nil
		},
	}
}
