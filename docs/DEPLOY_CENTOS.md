# CentOS 部署方案

## 服务器要求

- **系统**: CentOS 7/8 或 Rocky Linux 8/9
- **配置**: 2核4G 起步（推荐 4核8G）
- **端口**: 80 (HTTP)、443 (HTTPS)、3307 (MySQL，如需外网访问)

---

## 一、环境准备

### 1.1 更新系统

```bash
# CentOS 7
sudo yum update -y

# CentOS 8 / Rocky Linux
sudo dnf update -y
```

### 1.2 安装基础工具

```bash
sudo yum install -y git wget curl vim unzip epel-release
```

### 1.3 安装 MySQL 5.7

```bash
# 添加 MySQL 仓库
sudo yum localinstall -y https://dev.mysql.com/get/mysql57-community-release-el7-11.noarch.rpm

# 安装 MySQL
sudo yum install -y mysql-community-server

# 启动 MySQL
sudo systemctl start mysqld
sudo systemctl enable mysqld

# 获取临时密码
sudo grep 'temporary password' /var/log/mysqld.log

# 安全配置（修改密码、移除匿名用户等）
sudo mysql_secure_installation
```

### 1.4 创建数据库和用户

```bash
mysql -u root -p

# 在 MySQL 中执行：
CREATE DATABASE knowledge_base CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'kb_user'@'%' IDENTIFIED BY 'YourStrongPassword123!';
GRANT ALL PRIVILEGES ON knowledge_base.* TO 'kb_user'@'%';
FLUSH PRIVILEGES;
EXIT;
```

### 1.5 安装 Go（可选，用于编译）

```bash
# 下载 Go 1.21+
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz

# 解压到 /usr/local
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# 配置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh
source /etc/profile.d/go.sh

# 验证
go version
```

### 1.6 安装 Node.js（可选，用于前端构建）

```bash
# 安装 Node.js 18 LTS
curl -fsSL https://rpm.nodesource.com/setup_18.x | sudo bash -
sudo yum install -y nodejs

# 验证
node -v
npm -v
```

---

## 二、部署后端

### 2.1 上传代码

```bash
# 创建应用目录
sudo mkdir -p /opt/knowledge-base
sudo chown $USER:$USER /opt/knowledge-base

# 克隆代码（或使用 scp 上传）
cd /opt
git clone <your-repo-url> knowledge-base
cd knowledge-base/backend
```

### 2.2 配置环境变量

```bash
# 创建 .env 文件
cat > /opt/knowledge-base/backend/.env << 'EOF'
# 数据库配置
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=kb_user
DB_PASSWORD=YourStrongPassword123!
DB_NAME=knowledge_base

# JWT 密钥（请修改为随机字符串）
JWT_SECRET=your-random-jwt-secret-key-at-least-32-chars

# 服务配置
SERVER_PORT=8080
GIN_MODE=release
EOF

# 设置权限（仅所有者可读）
chmod 600 /opt/knowledge-base/backend/.env
```

### 2.3 编译后端

**方式一：在服务器编译**

```bash
cd /opt/knowledge-base/backend

# 配置 Go 代理（国内）
go env -w GOPROXY=https://goproxy.cn,direct

# 下载依赖
go mod download

# 编译
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o knowledge-base main.go
```

**方式二：本地编译后上传**

```powershell
# 在 Windows 本地
cd C:\Users\Administrator\.qclaw\workspace-agent-40338b03\knowledge-base\backend
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"
go build -o knowledge-base main.go

# 上传到服务器
scp knowledge-base user@your-server:/opt/knowledge-base/backend/
```

### 2.4 创建 systemd 服务

```bash
sudo tee /etc/systemd/system/knowledge-base.service << 'EOF'
[Unit]
Description=Knowledge Base API Server
After=network.target mysql.service

[Service]
Type=simple
User=nobody
Group=nobody
WorkingDirectory=/opt/knowledge-base/backend
ExecStart=/opt/knowledge-base/backend/knowledge-base
Restart=always
RestartSec=5
LimitNOFILE=65536

# 环境变量
EnvironmentFile=/opt/knowledge-base/backend/.env

# 安全配置
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

# 重载配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start knowledge-base
sudo systemctl enable knowledge-base

# 查看状态
sudo systemctl status knowledge-base

# 查看日志
sudo journalctl -u knowledge-base -f
```

---

## 三、部署前端

### 3.1 构建前端

**方式一：服务器构建**

```bash
cd /opt/knowledge-base/frontend

# 安装依赖
npm install

# 配置 API 地址
export VITE_API_BASE_URL=https://your-domain.com/api

# 构建
npm run build
```

**方式二：本地构建后上传**

```powershell
# 在 Windows 本地
cd C:\Users\Administrator\.qclaw\workspace-agent-40338b03\knowledge-base\frontend
$env:VITE_API_BASE_URL="https://your-domain.com/api"
npm run build

# 上传 dist 目录
scp -r dist/* user@your-server:/var/www/knowledge-base/
```

### 3.2 安装 Nginx

```bash
sudo yum install -y nginx
sudo systemctl start nginx
sudo systemctl enable nginx
```

### 3.3 配置 Nginx

```bash
sudo tee /etc/nginx/conf.d/knowledge-base.conf << 'EOF'
# HTTP 重定向到 HTTPS
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

# HTTPS
server {
    listen 443 ssl http2;
    server_name your-domain.com;

    # SSL 证书（使用 Let's Encrypt）
    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;

    # 前端静态文件
    root /var/www/knowledge-base;
    index index.html;

    # 前端路由（SPA）
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API 反向代理
    location /api {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        proxy_read_timeout 300s;
        proxy_connect_timeout 75s;
    }

    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    # Gzip 压缩
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml;
    gzip_min_length 1000;
}
EOF

# 测试配置
sudo nginx -t

# 重载 Nginx
sudo systemctl reload nginx
```

---

## 四、SSL 证书配置

### 4.1 安装 Certbot

```bash
# CentOS 7
sudo yum install -y certbot python2-certbot-nginx

# CentOS 8 / Rocky Linux
sudo dnf install -y certbot python3-certbot-nginx
```

### 4.2 申请证书

```bash
# 申请证书（Nginx 会自动配置）
sudo certbot --nginx -d your-domain.com

# 自动续期测试
sudo certbot renew --dry-run
```

---

## 五、防火墙配置

```bash
# 开放端口
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload

# 查看状态
sudo firewall-cmd --list-all
```

---

## 六、维护命令

### 6.1 后端管理

```bash
# 查看状态
sudo systemctl status knowledge-base

# 重启服务
sudo systemctl restart knowledge-base

# 查看日志
sudo journalctl -u knowledge-base -f --lines 100

# 更新部署
cd /opt/knowledge-base/backend
git pull
go build -o knowledge-base main.go
sudo systemctl restart knowledge-base
```

### 6.2 前端更新

```bash
cd /opt/knowledge-base/frontend
git pull
npm install
npm run build
sudo cp -r dist/* /var/www/knowledge-base/
```

---

## 七、数据库备份

### 7.1 手动备份

```bash
mysqldump -u kb_user -p knowledge_base > backup_$(date +%Y%m%d_%H%M%S).sql
```

### 7.2 定时备份（Cron）

```bash
# 创建备份脚本
sudo tee /opt/scripts/backup_db.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/backups/mysql"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

mysqldump -u kb_user -p'YourStrongPassword123!' knowledge_base | gzip > $BACKUP_DIR/kb_$DATE.sql.gz

# 保留最近 7 天
find $BACKUP_DIR -name "kb_*.sql.gz" -mtime +7 -delete
EOF

chmod +x /opt/scripts/backup_db.sh

# 添加定时任务（每天凌晨 3 点）
echo "0 3 * * * /opt/scripts/backup_db.sh" | crontab -
```

---

## 八、安全加固

### 8.1 SSH 安全

```bash
sudo vim /etc/ssh/sshd_config

# 修改以下配置：
Port 2222                    # 修改默认端口
PermitRootLogin no           # 禁止 root 登录
PasswordAuthentication no    # 强制密钥登录

sudo systemctl restart sshd
```

### 8.2 Fail2ban（防暴力破解）

```bash
sudo yum install -y fail2ban
sudo systemctl enable fail2ban
sudo systemctl start fail2ban
```

---

## 九、部署检查清单

- [ ] MySQL 已安装并创建数据库
- [ ] 后端服务运行正常（systemctl status）
- [ ] Nginx 配置正确并启动
- [ ] SSL 证书已申请
- [ ] 防火墙端口已开放
- [ ] 数据库备份脚本已配置
- [ ] 域名 DNS 解析已配置

---

## 十、常见问题

### Q1: 后端启动失败，数据库连接错误

检查 `.env` 配置和 MySQL 服务状态：
```bash
sudo systemctl status mysqld
mysql -u kb_user -p -h 127.0.0.1 knowledge_base
```

### Q2: 前端页面空白

检查 Nginx 日志：
```bash
sudo tail -f /var/log/nginx/error.log
```

### Q3: API 请求 502 Bad Gateway

检查后端服务是否运行：
```bash
sudo systemctl status knowledge-base
curl http://127.0.0.1:8080/api/categories
```
