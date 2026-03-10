package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wsafight/agent-lark/internal/auth"
	"github.com/wsafight/agent-lark/internal/cmdutil"
)

//go:embed SKILL.md
var skillMD []byte

func newInitCommand() *cobra.Command {
	var skillDir string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "初始化：安装 Skill 并配置凭据（新手入口）",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := cmdutil.GetGlobalFlags(cmd)
			yes, _ := cmd.Root().PersistentFlags().GetBool("yes")

			// 解析项目根目录，派生隔离 profile 名
			cwd, err := os.Getwd()
			if err != nil {
				cwd = "."
			}
			projectRoot := auth.DetectProjectRoot(cwd)
			if g.Profile == "" {
				g.Profile = auth.ProjectHashProfile(projectRoot)
			}

			// ── 检测已有状态 ──────────────────────────────────────
			target := filepath.Join(skillDir, "SKILL.md")
			skillExists := fileExists(target)

			var currentAppID string
			if cfg, _, err := auth.Load(g.Config, g.Profile); err == nil {
				currentAppID = cfg.AppID
			}
			credsConfigured := currentAppID != ""

			if skillExists || credsConfigured {
				fmt.Println("检测到已初始化：")
				if skillExists {
					fmt.Printf("  ✓ Skill 已存在        %s\n", target)
				}
				if credsConfigured {
					profilePath := auth.ProfileConfigPath(g.Profile)
					fmt.Printf("  ✓ 凭据已配置          App ID: %s（%s）\n", currentAppID, profilePath)
				}
				fmt.Println()

				if !yes {
					fmt.Print("是否重新初始化？[y/N]: ")
					reader := bufio.NewReader(os.Stdin)
					answer, _ := reader.ReadString('\n')
					if strings.TrimSpace(strings.ToLower(answer)) != "y" {
						fmt.Println("已取消。运行 agent-lark auth status 查看当前配置。")
						return nil
					}
					fmt.Println()
				}
			}

			// ── Step 1: 写入 SKILL.md ────────────────────────────
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			if err := os.WriteFile(target, skillMD, 0644); err != nil {
				return err
			}
			fmt.Printf("✓ Skill 已安装至 %s\n\n", target)

			// ── Step 2: 配置向导 ─────────────────────────────────
			fmt.Println("--- 配置凭据 ---")
			if _, err := runSetupWizard(g.Config, g.Profile); err != nil {
				return err
			}

			profilePath := auth.ProfileConfigPath(g.Profile)
			fmt.Printf("\n✓ 初始化完成！凭据存储于 %s\n", profilePath)
			fmt.Println("  运行 agent-lark auth status 查看配置详情。")
			return nil
		},
	}
	cmd.Flags().StringVar(&skillDir, "skill-dir", ".claude/skills/agent-lark", "Skill 安装目录")
	return cmd
}


func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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
