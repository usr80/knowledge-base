#!/bin/bash
# ============================================
# 一键构建 + Docker 打包脚本
# 使用：bash build-and-deploy.sh
# ============================================

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 配置
IMAGE_NAME="knowledge-base"
VERSION="${VERSION:-$(date +%Y%m%d-%H%M%S)}"
OUTPUT_DIR="./dist"

echo "=========================================="
echo "  知识库系统 - 构建与打包"
echo "=========================================="
echo "版本: ${VERSION}"
echo "输出: ${OUTPUT_DIR}"
echo ""

# 1. 清理旧文件
log_info "清理旧构建文件..."
rm -rf ${OUTPUT_DIR}
mkdir -p ${OUTPUT_DIR}

# 2. 编译后端
log_info "编译后端..."
cd backend
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
export GOPROXY=https://goproxy.cn,direct
go build -ldflags="-s -w" -o ../${OUTPUT_DIR}/knowledge-base main.go
cd ..
log_info "后端编译完成: ${OUTPUT_DIR}/knowledge-base"

# 3. 构建前端
log_info "构建前端..."
cd frontend
npm install --registry=https://registry.npmmirror.com
npm run build
cp -r dist ../${OUTPUT_DIR}/static
cd ..
log_info "前端构建完成: ${OUTPUT_DIR}/static"

# 4. 复制 Dockerfile
cp Dockerfile.deploy ${OUTPUT_DIR}/Dockerfile

# 5. 构建 Docker 镜像
log_info "构建 Docker 镜像..."
cd ${OUTPUT_DIR}
docker build -t ${IMAGE_NAME}:${VERSION} -t ${IMAGE_NAME}:latest .
cd ..
log_info "Docker 镜像构建完成: ${IMAGE_NAME}:${VERSION}"

# 6. 显示镜像信息
echo ""
echo "=========================================="
echo "  构建完成"
echo "=========================================="
docker images ${IMAGE_NAME}
echo ""
echo "运行命令："
echo "  docker run -d -p 8080:8080 \\"
echo "    -e DB_HOST=数据库地址 \\"
echo "    -e DB_PASSWORD=密码 \\"
echo "    -e JWT_SECRET=密钥 \\"
echo "    ${IMAGE_NAME}:${VERSION}"
echo ""
echo "或使用 docker-compose："
echo "  docker-compose -f docker-compose.deploy.yml up -d"
