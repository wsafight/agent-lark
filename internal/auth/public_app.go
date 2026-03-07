package auth

// publicAppID 和 publicAppSecret 在构建时通过 -ldflags 注入：
//
//	go build -ldflags "-X github.com/wangshian/agent-lark/internal/auth.publicAppID=cli_xxx \
//	                   -X github.com/wangshian/agent-lark/internal/auth.publicAppSecret=yyy" \
//	         ./cmd/agent-lark
//
// 开发本地构建时保持空值，命令会提示使用 agent-lark setup 替代。
var (
	publicAppID     string
	publicAppSecret string
)

// PublicAppAvailable 检查是否内置了公共应用凭据。
func PublicAppAvailable() bool {
	return publicAppID != "" && publicAppSecret != ""
}

// PublicAppCredentials 返回内置公共应用的 App ID 和 Secret。
func PublicAppCredentials() (appID, appSecret string) {
	return publicAppID, publicAppSecret
}
