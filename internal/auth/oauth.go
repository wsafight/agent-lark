package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkauthen "github.com/larksuite/oapi-sdk-go/v3/service/authen/v1"
)

const defaultOAuthPort = 9999

func OAuthLogin(appID, appSecret, scope, explicitConfig, profile string, port int) error {
	cfg, _, err := Load(explicitConfig, profile)
	if err != nil {
		return fmt.Errorf("AUTH_REQUIRED：请先执行 'agent-lark login'（或 'agent-lark auth login'）")
	}

	if port <= 0 {
		port = defaultOAuthPort
	}

	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)
	authDomain := openAuthDomain(cfg.Domain)

	authURLObj := &url.URL{
		Scheme: "https",
		Host:   authDomain,
		Path:   path.Join("/open-apis", "authen/v1/authorize"),
	}
	q := authURLObj.Query()
	q.Set("app_id", appID)
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", scope)
	authURLObj.RawQuery = q.Encode()
	authURL := authURLObj.String()

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	mux := http.NewServeMux()
	srv := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", port), Handler: mux}
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		codeCh <- code
		fmt.Fprintf(w, "<html><body><h2>授权成功！可以关闭此页面。</h2></body></html>")
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	fmt.Println("正在打开浏览器进行飞书授权...")
	fmt.Println("如果浏览器未自动打开，请手动访问：")
	fmt.Println(" ", authURL)
	openBrowser(authURL)

	const timeout = 2 * time.Minute
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var code string
	waiting := true
	for waiting {
		select {
		case code = <-codeCh:
			waiting = false
		case err := <-errCh:
			return fmt.Errorf("本地回调服务器启动失败（端口 %d 可能被占用）：%w\n可用 --port 指定其他端口重试", port, err)
		case t := <-ticker.C:
			remaining := time.Until(deadline).Round(time.Second)
			if remaining <= 0 {
				return fmt.Errorf("OAuth 超时：%s 内未完成授权\n请重新运行 'agent-lark auth oauth'", timeout)
			}
			fmt.Printf("\r⏳ 等待授权中... 剩余 %s   ", remaining)
			_ = t
		}
	}
	fmt.Println("\r✓ 授权回调已收到              ")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)

	clientOpts := []lark.ClientOptionFunc{}
	if cfg.Domain != "" {
		clientOpts = append(clientOpts, lark.WithOpenBaseUrl(cfg.Domain))
	}
	client := lark.NewClient(appID, appSecret, clientOpts...)
	req := larkauthen.NewCreateAccessTokenReqBuilder().
		Body(larkauthen.NewCreateAccessTokenReqBodyBuilder().
			GrantType("authorization_code").
			Code(code).
			Build()).
		Build()

	resp, err := client.Authen.AccessToken.Create(context.Background(), req)
	if err != nil {
		return fmt.Errorf("Token 换取失败：%w", err)
	}
	if !resp.Success() {
		return fmt.Errorf("Token 换取错误：%s（code %d）", resp.Msg, resp.Code)
	}

	cfg.UserSession = &UserSession{
		OpenID:          *resp.Data.OpenId,
		Name:            *resp.Data.Name,
		UserAccessToken: *resp.Data.AccessToken,
		RefreshToken:    *resp.Data.RefreshToken,
		ExpiresAt:       time.Now().Add(time.Duration(*resp.Data.ExpiresIn) * time.Second).Format(time.RFC3339),
	}
	if resp.Data.RefreshExpiresIn != nil {
		cfg.UserSession.RefreshExpiresAt = time.Now().Add(time.Duration(*resp.Data.RefreshExpiresIn) * time.Second).Format(time.RFC3339)
	}

	if err := Save(cfg, explicitConfig, profile); err != nil {
		return fmt.Errorf("保存配置失败：%w", err)
	}

	fmt.Printf("✓ 已授权为 %s (%s)\n", *resp.Data.Name, *resp.Data.OpenId)
	return nil
}

func openAuthDomain(openBaseDomain string) string {
	if strings.Contains(openBaseDomain, "larksuite") {
		return "open.larksuite.com"
	}
	return "open.feishu.cn"
}

// EnsureUserTokenValid refreshes user token when expired/almost expired.
func EnsureUserTokenValid(cfg *Config, explicitConfig, profile string) error {
	if cfg == nil || cfg.UserSession == nil || cfg.UserSession.UserAccessToken == "" {
		return fmt.Errorf("TOKEN_EXPIRED：用户 Token 不存在或已过期，请运行 'agent-lark auth oauth'")
	}
	if !isNearExpiry(cfg.UserSession.ExpiresAt, 30*time.Second) {
		return nil
	}
	if cfg.UserSession.RefreshToken == "" {
		return fmt.Errorf("TOKEN_EXPIRED：用户 Token 已过期，请运行 'agent-lark auth oauth'")
	}
	if err := refreshUserToken(cfg); err != nil {
		return err
	}
	if err := Save(cfg, explicitConfig, profile); err != nil {
		return fmt.Errorf("TOKEN_EXPIRED：用户 Token 刷新后保存失败：%w", err)
	}
	return nil
}

func refreshUserToken(cfg *Config) error {
	clientOpts := []lark.ClientOptionFunc{}
	if cfg.Domain != "" {
		clientOpts = append(clientOpts, lark.WithOpenBaseUrl(cfg.Domain))
	}
	client := lark.NewClient(cfg.AppID, cfg.AppSecret, clientOpts...)

	req := larkauthen.NewCreateRefreshAccessTokenReqBuilder().
		Body(larkauthen.NewCreateRefreshAccessTokenReqBodyBuilder().
			GrantType("refresh_token").
			RefreshToken(cfg.UserSession.RefreshToken).
			Build()).
		Build()

	resp, err := client.Authen.RefreshAccessToken.Create(context.Background(), req)
	if err != nil {
		return fmt.Errorf("TOKEN_EXPIRED：用户 Token 刷新失败：%w", err)
	}
	if !resp.Success() {
		return fmt.Errorf("TOKEN_EXPIRED：用户 Token 刷新失败：[%d] %s", resp.Code, resp.Msg)
	}
	if resp.Data == nil || resp.Data.AccessToken == nil || *resp.Data.AccessToken == "" {
		return fmt.Errorf("TOKEN_EXPIRED：用户 Token 刷新失败：响应缺少 access_token")
	}

	cfg.UserSession.UserAccessToken = *resp.Data.AccessToken
	if resp.Data.RefreshToken != nil && *resp.Data.RefreshToken != "" {
		cfg.UserSession.RefreshToken = *resp.Data.RefreshToken
	}
	if resp.Data.ExpiresIn != nil {
		cfg.UserSession.ExpiresAt = time.Now().Add(time.Duration(*resp.Data.ExpiresIn) * time.Second).Format(time.RFC3339)
	}
	if resp.Data.RefreshExpiresIn != nil {
		cfg.UserSession.RefreshExpiresAt = time.Now().Add(time.Duration(*resp.Data.RefreshExpiresIn) * time.Second).Format(time.RFC3339)
	}
	if resp.Data.OpenId != nil && *resp.Data.OpenId != "" {
		cfg.UserSession.OpenID = *resp.Data.OpenId
	}
	if resp.Data.Name != nil && *resp.Data.Name != "" {
		cfg.UserSession.Name = *resp.Data.Name
	}

	return nil
}

func isNearExpiry(expiresAt string, threshold time.Duration) bool {
	t, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return true
	}
	return time.Until(t) <= threshold
}

func openBrowser(url string) {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
		url = "url.dll,FileProtocolHandler " + url
	default:
		cmd = "xdg-open"
	}
	_ = exec.Command(cmd, url).Start()
}
