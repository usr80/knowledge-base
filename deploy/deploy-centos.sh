#!/bin/bash
# CentOS 一键部署脚本
# 使用：sudo bash deploy.sh

set -e

# ===== 配置区域 =====
DB_NAME="knowledge_base"
DB_USER="kb_user"
DB_PASS="ChangeThisPassword123!"
JWT_SECRET="change-this-to-random-32-chars-string"
DOMAIN="your-domain.com"
# ====================

echo "=========================================="
echo "  知识库系统 CentOS 部署脚本"
echo "=========================================="

# 检查 root 权限
if [ "$EUID" -ne 0 ]; then
  echo "请使用 sudo 运行此脚本"
  exit 1
fi

# 1. 安装依赖
echo "[1/8] 安装系统依赖..."
yum install -y git wget curl vim unzip epel-release

# 2. 安装 MySQL
echo "[2/8] 安装 MySQL 5.7..."
if ! command -v mysql &> /dev/null; then
  yum localinstall -y https://dev.mysql.com/get/mysql57-community-release-el7-11.noarch.rpm
  yum install -y mysql-community-server
  systemctl start mysqld
  systemctl enable mysqld
fi

# 3. 创建数据库
echo "[3/8] 配置数据库..."
mysql -u root << EOF
CREATE DATABASE IF NOT EXISTS ${DB_NAME} CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS '${DB_USER}'@'localhost' IDENTIFIED BY '${DB_PASS}';
GRANT ALL PRIVILEGES ON ${DB_NAME}.* TO '${DB_USER}'@'localhost';
FLUSH PRIVILEGES;
EOF

# 4. 安装 Go
echo "[4/8] 安装 Go..."
if ! command -v go &> /dev/null; then
  wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
  rm -rf /usr/local/go
  tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
  echo 'export PATH=$PATH:/usr/local/go/bin' > /etc/profile.d/go.sh
  source /etc/profile.d/go.sh
fi

# 5. 安装 Node.js
echo "[5/8] 安装 Node.js..."
if ! command -v node &> /dev/null; then
  curl -fsSL https://rpm.nodesource.com/setup_18.x | bash -
  yum install -y nodejs
fi

# 6. 安装 Nginx
echo "[6/8] 安装 Nginx..."
if ! command -v nginx &> /dev/null; then
  yum install -y nginx
  systemctl start nginx
  systemctl enable nginx
fi

# 7. 配置应用目录
echo "[7/8] 配置应用..."
mkdir -p /opt/knowledge-base
mkdir -p /var/www/knowledge-base
mkdir -p /opt/backups/mysql

# 创建 .env 文件
cat > /opt/knowledge-base/backend/.env << EOF
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=${DB_USER}
DB_PASSWORD=${DB_PASS}
DB_NAME=${DB_NAME}
JWT_SECRET=${JWT_SECRET}
SERVER_PORT=8080
GIN_MODE=release
EOF

chmod 600 /opt/knowledge-base/backend/.env

# 8. 配置防火墙
echo "[8/8] 配置防火墙..."
firewall-cmd --permanent --add-service=http
firewall-cmd --permanent --add-service=https
firewall-cmd --reload

echo ""
echo "=========================================="
echo "  基础环境安装完成！"
echo "=========================================="
echo ""
echo "后续步骤："
echo "1. 上传代码到 /opt/knowledge-base/"
echo "2. 编译后端："
echo "   cd /opt/knowledge-base/backend"
echo "   go env -w GOPROXY=https://goproxy.cn,direct"
echo "   CGO_ENABLED=0 go build -o knowledge-base main.go"
echo ""
echo "3. 构建前端："
echo "   cd /opt/knowledge-base/frontend"
echo "   npm install"
echo "   npm run build"
echo "   cp -r dist/* /var/www/knowledge-base/"
echo ""
echo "4. 配置 Nginx（参考 docs/DEPLOY_CENTOS.md）"
echo "5. 启动后端服务："
echo "   systemctl start knowledge-base"
echo ""
echo "数据库信息："
echo "  数据库：${DB_NAME}"
echo "  用户名：${DB_USER}"
echo "  密码：${DB_PASS}"
echo ""
