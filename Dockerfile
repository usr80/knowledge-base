# ============================================
# 多阶段构建：前端 + 后端 → 单镜像
# ============================================

# ---------- 阶段1：构建前端 ----------
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

# 先复制依赖文件，利用 Docker 缓存
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install --registry=https://registry.npmmirror.com

# 复制前端源码并构建
COPY frontend/ ./
RUN npm run build

# ---------- 阶段2：构建后端 ----------
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app/backend

# 先复制依赖文件，利用 Docker 缓存
COPY backend/go.mod backend/go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download

# 复制后端源码并编译
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o knowledge-base main.go

# ---------- 阶段3：最终镜像 ----------
FROM alpine:3.19

# 安装时区数据和 ca 证书
RUN apk add --no-cache tzdata ca-certificates \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && apk del tzdata

WORKDIR /app

# 从 backend-builder 复制二进制文件
COPY --from=backend-builder /app/backend/knowledge-base .

# 从 frontend-builder 复制前端构建产物
COPY --from=frontend-builder /app/frontend/dist ./static

# 复制配置模板（实际配置通过环境变量或挂载 .env 提供）
COPY backend/.env.example ./.env.example

# 暴露端口
EXPOSE 8080

# 环境变量（可通过 docker run -e 或 docker-compose 覆盖）
ENV STATIC_DIR=/app/static
ENV GIN_MODE=release
ENV SERVER_PORT=8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

# 启动
CMD ["./knowledge-base"]
