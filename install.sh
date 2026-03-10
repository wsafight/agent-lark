#!/usr/bin/env bash
set -euo pipefail

REPO="wsafight/agent-lark"
BINARY="agent-lark"

# ── 颜色输出 ────────────────────────────────────────────────
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'
info()  { echo -e "${GREEN}✓${NC} $*"; }
warn()  { echo -e "${YELLOW}⚠${NC} $*"; }
error() { echo -e "${RED}✗${NC} $*" >&2; exit 1; }

# ── 检测系统 ─────────────────────────────────────────────────
case "$(uname -s)" in
  Darwin) OS="darwin" ;;
  Linux)  OS="linux"  ;;
  *)      error "不支持的操作系统: $(uname -s)" ;;
esac

case "$(uname -m)" in
  x86_64)        ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)             error "不支持的架构: $(uname -m)" ;;
esac

# ── 获取版本 ─────────────────────────────────────────────────
# 支持 VERSION=v0.2.0 ./install.sh 固定版本
if [ -z "${VERSION:-}" ]; then
  echo "正在获取最新版本..."
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | sed -E 's/.*"([^"]+)".*/\1/')
fi

[ -z "$VERSION" ] && error "无法获取版本信息，请检查网络连接"

# ── 下载二进制 ───────────────────────────────────────────────
ASSET="${BINARY}_${OS}_${ARCH}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"

echo "正在下载 ${BINARY} ${VERSION} (${OS}/${ARCH})..."
TMP=$(mktemp)
trap 'rm -f "$TMP"' EXIT

if ! curl -fsSL --progress-bar -o "$TMP" "$URL"; then
  error "下载失败，请检查网络或访问 https://github.com/${REPO}/releases 手动下载"
fi

chmod +x "$TMP"

# ── 安装 ─────────────────────────────────────────────────────
INSTALL_DIR="/usr/local/bin"

install_binary() {
  if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP" "${INSTALL_DIR}/${BINARY}"
  elif command -v sudo &>/dev/null; then
    echo "需要管理员权限安装到 ${INSTALL_DIR}..."
    sudo mv "$TMP" "${INSTALL_DIR}/${BINARY}"
  else
    # 降级到用户目录
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
    mv "$TMP" "${INSTALL_DIR}/${BINARY}"
    warn "已安装到 ${INSTALL_DIR}"
    warn "请将以下内容添加到 ~/.bashrc 或 ~/.zshrc："
    warn "  export PATH=\"\$PATH:${HOME}/.local/bin\""
  fi
}

install_binary

# ── 验证 ─────────────────────────────────────────────────────
if command -v "$BINARY" &>/dev/null || [ -x "${INSTALL_DIR}/${BINARY}" ]; then
  info "${BINARY} ${VERSION} 已安装到 ${INSTALL_DIR}/${BINARY}"
else
  warn "安装完成，但 ${BINARY} 不在 PATH 中，请手动配置"
fi

echo ""
echo "快速开始："
echo "  agent-lark init      # 安装 Skill 并配置凭据"
echo "  agent-lark doctor    # 诊断配置问题"
echo "  agent-lark --help    # 查看所有命令"
