package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

//go:embed SKILL.md
var skillMD []byte

func newInstallCommand() *cobra.Command {
	var skillDir string
	cmd := &cobra.Command{
		Use:   "install",
		Short: "将 Agent Skill 安装到当前 Claude Code 项目",
		RunE: func(cmd *cobra.Command, args []string) error {
			target := filepath.Join(skillDir, "SKILL.md")
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			if err := os.WriteFile(target, skillMD, 0644); err != nil {
				return err
			}
			fmt.Printf("✓ Skill 已安装至 %s\n", target)
			return nil
		},
	}
	cmd.Flags().StringVar(&skillDir, "skill-dir", ".claude/skills/agent-lark", "Skill 安装目录")
	return cmd
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "打印版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("agent-lark %s\n  oapi-sdk-go: v3.5.3\n  go: go1.22\n", Version)
		},
	}
}

// Version 通过 ldflags 注入，默认 dev。
var Version = "dev"
