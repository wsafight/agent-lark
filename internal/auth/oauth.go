package auth

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkauthen "github.com/larksuite/oapi-sdk-go/v3/service/authen/v1"
)

const redirectURI = "http://127.0.0.1:9999/callback"

func OAuthLogin(appID, appSecret, scope, explicitConfig, profile string) error {
	cfg, _, err := Load(explicitConfig, profile)
	if err != nil {
		return fmt.Errorf("AUTH_REQUIRED：请先执行 'agent-lark login'（或 'agent-lark auth login'）")
	}

	authURL := fmt.Sprintf(
		"https://open.feishu.cn/open-apis/authen/v1/authorize?app_id=%s&redirect_uri=%s&scope=%s",
		appID, redirectURI, scope,
	)

	codeCh := make(chan string, 1)
	mux := http.NewServeMux()
	srv := &http.Server{Addr: "127.0.0.1:9999", Handler: mux}
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		codeCh <- code
		fmt.Fprintf(w, "<html><body><h2>授权成功！可以关闭此页面。</h2></body></html>")
	})

	go func() { _ = srv.ListenAndServe() }()

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

	client := lark.NewClient(appID, appSecret)
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

	if err := Save(cfg, explicitConfig, profile); err != nil {
		return fmt.Errorf("保存配置失败：%w", err)
	}

	fmt.Printf("✓ 已授权为 %s (%s)\n", *resp.Data.Name, *resp.Data.OpenId)
	return nil
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
