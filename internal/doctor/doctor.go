package doctor

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/spf13/cobra"
	"github.com/wangshian/agent-lark/internal/auth"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "诊断配置问题",
		RunE:  runDoctor,
	}
}

func runDoctor(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Root().PersistentFlags().GetString("profile")
	config, _ := cmd.Root().PersistentFlags().GetString("config")

	ok := func(msg string) { fmt.Printf("✓ %s\n", msg) }
	fail := func(msg, hint string) { fmt.Printf("✗ %s\n  → %s\n", msg, hint) }

	// 1. 检查配置文件
	cfgPath := auth.ResolveConfigPath(config, profile)
	if _, err := os.Stat(cfgPath); err != nil {
		fail("配置文件不存在  "+cfgPath, "运行: agent-lark login 或 agent-lark setup")
		return nil
	}
	ok("配置文件存在          " + cfgPath)

	// 2. 加载并验证配置
	cfg, _, err := auth.Load(config, profile)
	if err != nil {
		fail("配置文件损坏或无法读取", err.Error())
		return nil
	}
	if cfg.AppID == "" || cfg.AppSecret == "" {
		fail("应用凭据缺失", "运行: agent-lark setup")
		return nil
	}

	// 3. 验证应用凭据
	client := lark.NewClient(cfg.AppID, cfg.AppSecret)
	resp, err := client.Auth.TenantAccessToken.Internal(context.Background(), nil)
	if err != nil || !resp.Success() {
		fail("应用凭据无效  "+cfg.AppID, "检查 App ID 和 App Secret 是否正确")
		return nil
	}
	ok("应用凭据有效          " + cfg.AppID)

	// 4. 检查 API 连通性
	domain := "open.feishu.cn"
	if cfg.Domain != "" {
		domain = cfg.Domain
	}
	start := time.Now()
	httpResp, err := http.Get("https://" + domain + "/open-apis/auth/v3/tenant_access_token/internal")
	latency := time.Since(start)
	if err != nil || httpResp.StatusCode >= 500 {
		fail("API 连通性检查失败  "+domain, "检查网络连接")
	} else {
		ok(fmt.Sprintf("API 连通性正常        %s 延迟 %dms", domain, latency.Milliseconds()))
	}

	// 5. 检查用户 Token
	if cfg.UserSession == nil || cfg.UserSession.UserAccessToken == "" {
		fail("用户 Token 未配置", "运行: agent-lark auth oauth")
	} else {
		expiresAt, _ := time.Parse(time.RFC3339, cfg.UserSession.ExpiresAt)
		if time.Now().After(expiresAt) {
			fail("用户 Token 已过期", "运行: agent-lark auth oauth")
		} else {
			remaining := time.Until(expiresAt).Round(time.Minute)
			ok(fmt.Sprintf("用户 Token 有效        %s（剩余 %s）", cfg.UserSession.Name, remaining))
		}
	}

	return nil
}
