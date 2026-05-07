#!/bin/bash
# ============================================
# 一键构建脚本（Linux/macOS）
# 输出：build/backend/ + build/frontend/
# ============================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_step() { echo -e "${CYAN}[$1]${NC} $2"; }

echo "=========================================="
echo -e "${CYAN}  知识库系统 - 构建${NC}"
echo "=========================================="
echo "输出目录: $BUILD_DIR"
echo ""

# 1. 清理并创建 build 目录
log_step "1/4" "清理旧构建文件..."
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR/backend"
mkdir -p "$BUILD_DIR/frontend"

# 2. 编译后端
log_step "2/4" "编译后端..."
cd "$PROJECT_ROOT/backend"
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
export GOPROXY=https://goproxy.cn,direct
go build -ldflags="-s -w" -o "$BUILD_DIR/backend/knowledge-base" main.go
echo "  → build/backend/knowledge-base"

# 3. 构建前端
log_step "3/4" "构建前端..."
cd "$PROJECT_ROOT/frontend"
npm run build
echo "  → build/frontend/dist/"

# 4. 复制配置模板
log_step "4/4" "复制配置文件..."
cp "$PROJECT_ROOT/backend/.env.example" "$BUILD_DIR/backend/.env.example"

echo ""
echo "=========================================="
echo -e "${CYAN}  构建完成${NC}"
echo "=========================================="
echo ""
echo "构建产物："
find "$BUILD_DIR" -type f -exec ls -lh {} \; | awk '{print "  " $NF " (" $5 ")"}'
echo ""
echo -e "${YELLOW}部署步骤：${NC}"
echo "  1. 编辑 build/backend/.env（配置数据库密码等）"
echo "  2. docker-compose up -d"
