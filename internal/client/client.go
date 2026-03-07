package client

import (
	"fmt"
	"os"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/wangshian/agent-lark/internal/auth"
)

type Options struct {
	TokenMode string // "auto" | "tenant" | "user"
	Debug     bool
	Profile   string
	Config    string
	Domain    string
}

// Result 包含已构建的 Client 和生效的 Token 模式。
type Result struct {
	Client    *lark.Client
	Mode      string // "tenant" 或 "user"
	UserToken string // mode=="user" 时非空
	Cfg       *auth.Config
}

func New(opts Options) (*Result, error) {
	// 从环境变量读取（CI 场景）
	if envID := os.Getenv("LARK_APP_ID"); envID != "" {
		if envSecret := os.Getenv("LARK_APP_SECRET"); envSecret != "" {
			opts.Config = "" // 忽略文件配置
			cfg := &auth.Config{
				AppID:            envID,
				AppSecret:        envSecret,
				DefaultTokenMode: "tenant",
			}
			return buildResult(cfg, opts)
		}
	}

	cfg, _, err := auth.Load(opts.Config, opts.Profile)
	if err != nil {
		return nil, err
	}
	return buildResult(cfg, opts)
}

func buildResult(cfg *auth.Config, opts Options) (*Result, error) {
	clientOpts := []lark.ClientOptionFunc{}
	if opts.Debug {
		clientOpts = append(clientOpts, lark.WithLogLevel(larkcore.LogLevelDebug))
	}
	if opts.Domain != "" {
		clientOpts = append(clientOpts, lark.WithOpenBaseUrl(opts.Domain))
	} else if cfg.Domain != "" {
		clientOpts = append(clientOpts, lark.WithOpenBaseUrl(cfg.Domain))
	}

	client := lark.NewClient(cfg.AppID, cfg.AppSecret, clientOpts...)

	mode := opts.TokenMode
	if mode == "" || mode == "auto" {
		mode = cfg.DefaultTokenMode
	}
	if mode == "" {
		mode = "auto"
	}

	var userToken string
	if mode == "auto" || mode == "user" {
		if cfg.UserSession != nil && cfg.UserSession.UserAccessToken != "" {
			userToken = cfg.UserSession.UserAccessToken
			mode = "user"
		} else if mode == "user" {
			return nil, fmt.Errorf("TOKEN_EXPIRED：用户 Token 不存在或已过期，请运行 'agent-lark auth oauth'")
		} else {
			mode = "tenant"
		}
	}

	return &Result{Client: client, Mode: mode, UserToken: userToken, Cfg: cfg}, nil
}

// RequestOptions 根据 Token 模式返回每次 API 调用需要附带的请求选项。
func (r *Result) RequestOptions() []larkcore.RequestOptionFunc {
	if r.Mode == "user" && r.UserToken != "" {
		return []larkcore.RequestOptionFunc{
			larkcore.WithUserAccessToken(r.UserToken),
		}
	}
	return nil
}
