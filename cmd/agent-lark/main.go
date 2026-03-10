package main

import (
	"os"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/auth"
	"github.com/wsafight/agent-lark/internal/base"
	"github.com/wsafight/agent-lark/internal/comments"
	"github.com/wsafight/agent-lark/internal/contact"
	"github.com/wsafight/agent-lark/internal/docs"
	"github.com/wsafight/agent-lark/internal/doctor"
	"github.com/wsafight/agent-lark/internal/im"
	"github.com/wsafight/agent-lark/internal/output"
	"github.com/wsafight/agent-lark/internal/perms"
	"github.com/wsafight/agent-lark/internal/task"
	tmpl "github.com/wsafight/agent-lark/internal/template"
	"github.com/wsafight/agent-lark/internal/wiki"
)

func main() {
	root := &cobra.Command{
		Use:           "agent-lark",
		Short:         "飞书/Lark Agent CLI",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			agentMode, _ := cmd.Root().PersistentFlags().GetBool("agent")
			if agentMode {
				output.GlobalAgent = true
			}
		},
	}

	root.PersistentFlags().String("format", "text", "输出格式：text|json|md")
	root.PersistentFlags().String("token-mode", "auto", "Token 模式：auto|tenant|user")
	root.PersistentFlags().String("profile", "", "凭据 profile（留空时自动解析）")
	root.PersistentFlags().String("config", "", "显式指定凭据文件路径")
	root.PersistentFlags().String("domain", "", "覆盖 API 域名")
	root.PersistentFlags().Bool("debug", false, "开启调试输出")
	root.PersistentFlags().Bool("quiet", false, "静默模式，仅输出数据")
	root.PersistentFlags().Bool("agent", false, "Agent 模式：--format json --yes + 结构化错误")
	root.PersistentFlags().Bool("yes", false, "自动确认所有提示")

	root.AddCommand(
		newLoginCommand(),
		newSetupCommand(),
		NewAuthCommand(),
		docs.NewCommand(),
		wiki.NewCommand(),
		contact.NewCommand(),
		im.NewCommand(),
		comments.NewCommand(),
		base.NewCommand(),
		task.NewCommand(),
		doctor.NewCommand(),
		tmpl.NewCommand(),
		perms.NewCommand(),
		newInstallCommand(),
		newVersionCommand(),
	)

	_ = auth.ProfilesDir() // ensure import

	if err := root.Execute(); err != nil {
		code, message := splitErrorCode(err)
		output.PrintErrorCode(code, message, hintForCode(code))
		os.Exit(1)
	}
}

func splitErrorCode(err error) (string, string) {
	if err == nil {
		return "ERROR", ""
	}
	msg := strings.TrimSpace(err.Error())
	sep := "："
	idx := strings.Index(msg, sep)
	if idx < 0 {
		sep = ":"
		idx = strings.Index(msg, sep)
	}
	if idx <= 0 {
		return "ERROR", msg
	}
	code := strings.TrimSpace(msg[:idx])
	if !looksLikeErrorCode(code) {
		return "ERROR", msg
	}
	return code, strings.TrimSpace(msg[idx+len(sep):])
}

func looksLikeErrorCode(code string) bool {
	if code == "" {
		return false
	}
	for _, r := range code {
		if !(unicode.IsUpper(r) || unicode.IsDigit(r) || r == '_') {
			return false
		}
	}
	return true
}

func hintForCode(code string) string {
	switch code {
	case "AUTH_REQUIRED":
		return "运行: agent-lark login 或 agent-lark setup"
	case "TOKEN_EXPIRED":
		return "运行: agent-lark auth oauth"
	case "MISSING_FLAG", "MISSING_ARG":
		return "运行: agent-lark <命令> --help 查看参数说明"
	default:
		return ""
	}
}
