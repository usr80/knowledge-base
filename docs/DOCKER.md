# Docker 部署方案

## 架构

```
┌─────────────────────────────────────┐
│  单容器运行（前后端一体）            │
│                                     │
│  Go + Gin 后端                      │
│  ├─ /api/* → API 接口               │
│  └─ 其他 → 前端静态文件              │
│                                     │
│  Vue3 + Vite 前端                   │
│  ├─ 构建产物 /app/static            │
│  └─ SPA 路由返回 index.html         │
└─────────────────────────────────────┘
```

## 快速部署

### 方式一：Docker Compose（推荐）

```bash
# 创建环境变量文件
cat > .env << 'EOF'
DB_PASSWORD=YourStrongPassword123!
JWT_SECRET=your-random-secret-32chars
MYSQL_ROOT_PASSWORD=RootPassword123!
EOF

# 一键启动（应用 + MySQL）
docker-compose up -d

# 查看状态
docker-compose ps

# 查看日志
docker-compose logs -f app

# 访问
http://localhost:8080
```

### 方式二：单独运行（已有数据库）

```bash
# 构建镜像
docker build -t knowledge-base .

# 运行容器
docker run -d \
  --name knowledge-base \
  -p 8080:8080 \
  -e DB_HOST=192.168.1.100 \
  -e DB_PORT=3306 \
  -e DB_USER=kb_user \
  -e DB_PASSWORD=YourPassword \
  -e DB_NAME=knowledge_base \
  -e JWT_SECRET=your-random-secret \
  knowledge-base

# 查看日志
docker logs -f knowledge-base
```

---

## 配置方式

### 1. 环境变量（推荐）

```bash
docker run -d \
  -e DB_HOST=数据库地址 \
  -e DB_PASSWORD=密码 \
  -e JWT_SECRET=密钥 \
  knowledge-base
```

### 2. 挂载配置文件

```bash
# 创建 .env 文件
cat > config.env << 'EOF'
DB_HOST=192.168.1.100
DB_PASSWORD=YourPassword
JWT_SECRET=your-secret
GIN_MODE=release
EOF

# 挂载到容器
docker run -d \
  -v $(pwd)/config.env:/app/.env \
  knowledge-base

# 注意：需要修改 main.go 加载 /app/.env
```

### 3. docker-compose.yml 配置

```yaml
services:
  app:
    environment:
      - DB_HOST=db         # 使用 compose 中的 db 服务
      - DB_PASSWORD=${DB_PASSWORD}  # 从 .env 文件读取
      - JWT_SECRET=${JWT_SECRET}
```

---

## 数据持久化

### MySQL 数据卷

```yaml
volumes:
  mysql-data:
    driver: local
```

数据存储在 Docker 卷中，容器重启不丢失：

```bash
# 查看卷
docker volume ls

# 查看卷内容
docker volume inspect knowledge-base_mysql-data

# 备份
docker exec knowledge-base-mysql mysqldump -u root -p knowledge_base > backup.sql

# 恢复
docker exec -i knowledge-base-mysql mysql -u root -p knowledge_base < backup.sql
```

---

## 常用命令

```bash
# 启动
docker-compose up -d

# 停止
docker-compose down

# 重启
docker-compose restart app

# 查看日志
docker-compose logs -f app

# 进入容器
docker exec -it knowledge-base sh

# 重建镜像（代码更新后）
docker-compose build --no-cache app
docker-compose up -d app

# 清理
docker-compose down -v  # 删除容器和卷
```

---

## 生产环境部署

### 1. Nginx 反向代理（HTTPS）

```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 2. systemd 管理（可选）

```ini
[Unit]
Description=Knowledge Base Docker
After=docker.service

[Service]
Type=simple
ExecStart=/usr/bin/docker-compose -f /opt/knowledge-base/docker-compose.yml up
ExecStop=/usr/bin/docker-compose -f /opt/knowledge-base/docker-compose.yml down
Restart=always

[Install]
WantedBy=multi-user.target
```

### 3. 安全配置

```bash
# 生成随机密钥
openssl rand -base64 32

# .env 文件权限
chmod 600 .env
```

---

## Dockerfile 说明

### 多阶段构建

```dockerfile
# 阶段1：前端构建（node:18-alpine）
FROM node:18-alpine AS frontend-builder
→ npm run build → dist/

# 阶段2：后端编译（golang:1.21-alpine）
FROM golang:1.21-alpine AS backend-builder
→ go build → knowledge-base

# 阶段3：最终镜像（alpine:3.19）
FROM alpine:3.19
→ 只包含：二进制 + 静态文件 + 配置模板
→ 镜像大小：~20MB（不含 Node/Go）
```

### 镜像大小对比

| 方式 | 大小 |
|------|------|
| 多阶段构建 | ~20MB |
| 单阶段（含 Node+Go） | ~500MB |

---

## 常见问题

### Q1: 前端页面空白

检查静态文件是否正确复制：

```bash
docker exec -it knowledge-base ls -la /app/static
```

### Q2: API 请求 404

检查路由配置，确保 `/api` 路由优先于 SPA 路由。

### Q3: 数据库连接失败

检查 MySQL 容器状态和健康检查：

```bash
docker-compose ps
docker-compose logs db
```

### Q4: 更新代码后重新部署

```bash
# 拉取最新代码
git pull

# 重建并重启
docker-compose build --no-cache app
docker-compose up -d app
```

---

## 镜像推送（可选）

```bash
# 登录 Docker Hub
docker login

# 标记镜像
docker tag knowledge-base yourname/knowledge-base:latest

# 推送
docker push yourname/knowledge-base:latest

# 远程部署
docker pull yourname/knowledge-base:latest
docker-compose up -d
```