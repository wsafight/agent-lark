package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	githubRepo          = "wsafight/agent-lark"
	upgradeCheckInterval = 24 * time.Hour
)

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// ── 命令注册 ──────────────────────────────────────────────────

func newUpgradeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "升级到最新版本",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print("检查最新版本...")
			release, err := fetchLatestRelease()
			if err != nil {
				return fmt.Errorf("API_ERROR：获取版本信息失败：%w", err)
			}
			fmt.Printf(" %s\n", release.TagName)

			current := strings.TrimPrefix(Version, "v")
			latest := strings.TrimPrefix(release.TagName, "v")
			if current != "dev" && current == latest {
				fmt.Printf("✓ 已是最新版本 %s，无需升级\n", Version)
				return nil
			}

			// 显式调用：有进度输出，成功后提示重新打开
			if err := doUpgrade(release, true); err != nil {
				return err
			}
			fmt.Println("\n请重新运行命令以使用新版本。")
			return nil
		},
	}
}

// ── 静默后台升级 ──────────────────────────────────────────────

// StartBackgroundUpgrade 在后台静默检查并升级，每 24h 至多执行一次。
// 应在 PersistentPreRun 中调用；upgrade/version 命令自身跳过。
func StartBackgroundUpgrade() {
	if Version == "dev" {
		return
	}
	go func() {
		if !upgradeCheckDue() {
			return
		}
		touchUpgradeCheck() // 先标记，防止多进程并发请求

		release, err := fetchLatestRelease()
		if err != nil {
			return
		}

		current := strings.TrimPrefix(Version, "v")
		latest := strings.TrimPrefix(release.TagName, "v")
		if current == latest {
			return
		}

		_ = doUpgrade(release, false) // 静默，忽略错误
	}()
}

// ── 核心升级逻辑 ──────────────────────────────────────────────

// doUpgrade 下载并原子替换当前二进制。
// verbose=true 时打印进度；false 时完全静默。
// os.Rename 会删除旧版本文件（从目录项中移除），无需额外清理。
func doUpgrade(release *githubRelease, verbose bool) error {
	assetName := fmt.Sprintf("agent-lark_%s_%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		assetName += ".exe"
	}

	var downloadURL string
	for _, a := range release.Assets {
		if a.Name == assetName {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("UNSUPPORTED：未找到适用于 %s/%s 的二进制文件", runtime.GOOS, runtime.GOARCH)
	}

	// 当前二进制的真实路径（解析符号链接）
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("FILE_ERROR：无法获取当前程序路径：%w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("FILE_ERROR：无法解析程序路径：%w", err)
	}

	// 临时文件与目标在同一目录 → 同一文件系统 → Rename 原子成功
	tmpPath := execPath + ".new"

	if verbose {
		fmt.Printf("正在下载 %s (%s/%s)...", release.TagName, runtime.GOOS, runtime.GOARCH)
	}

	if err := downloadFile(downloadURL, tmpPath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("FILE_ERROR：设置权限失败：%w", err)
	}

	// Rename 原子替换：旧版本文件被移除（inode 在当前进程退出后释放）
	if err := os.Rename(tmpPath, execPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("FILE_ERROR：替换二进制失败（可能需要 sudo）：%w", err)
	}

	if verbose {
		fmt.Printf(" ✓\n✓ 已升级：%s → %s\n", Version, release.TagName)
	}
	return nil
}

// ── 升级检查频率控制 ──────────────────────────────────────────

func upgradeCheckFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".agent-lark", "upgrade-check")
}

func upgradeCheckDue() bool {
	info, err := os.Stat(upgradeCheckFilePath())
	return err != nil || time.Since(info.ModTime()) > upgradeCheckInterval
}

func touchUpgradeCheck() {
	path := upgradeCheckFilePath()
	_ = os.MkdirAll(filepath.Dir(path), 0700)
	_ = os.WriteFile(path, nil, 0600)
}

// ── HTTP 工具 ─────────────────────────────────────────────────

func fetchLatestRelease() (*githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var r githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return fmt.Errorf("API_ERROR：下载失败：%w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API_ERROR：下载失败，HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("FILE_ERROR：创建临时文件失败：%w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("FILE_ERROR：写入失败：%w", err)
	}
	return nil
}
