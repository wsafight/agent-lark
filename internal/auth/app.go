package auth

import (
	"context"
	"fmt"

	lark "github.com/larksuite/oapi-sdk-go/v3"
)

// ValidateAppCredentials 通过获取 tenant_access_token 验证应用凭据是否有效。
func ValidateAppCredentials(appID, appSecret string) error {
	client := lark.NewClient(appID, appSecret)
	// 调用一个简单 API 验证连通性
	resp, err := client.Auth.TenantAccessToken.Internal(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("API 连通性检查失败：%w", err)
	}
	if !resp.Success() {
		return fmt.Errorf("应用凭据无效：%s（code %d）", resp.Msg, resp.Code)
	}
	return nil
}
