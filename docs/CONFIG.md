# 配置管理说明

## 配置优先级

```
命令行参数 > 环境变量 > .env.{env} > .env > 默认值
```

## 配置文件说明

| 文件 | 用途 | Git 跟踪 |
|------|------|---------|
| `.env.example` | 配置模板 | ✅ 是 |
| `.env.dev` | 开发环境 | ❌ 否 |
| `.env.test` | 测试环境 | ❌ 否 |
| `.env.prod` | 生产环境模板 | ❌ 否 |
| `.env` | 当前环境配置 | ❌ 否 |
| `.env.local` | 本地覆盖配置 | ❌ 否 |

## 使用方式

### 1. 开发环境（默认）

```bash
# 方式一：使用 .env.dev
cp .env.dev .env
go run main.go

# 方式二：指定环境
go run main.go -e dev

# 方式三：显示配置（调试）
go run main.go -e dev -show-config
```

### 2. 测试环境

```bash
go run main.go -e test
# 或
cp .env.test .env
go run main.go
```

### 3. 生产环境

**方式一：配置文件**

```bash
# 上传配置文件到服务器
scp .env.prod user@server:/opt/knowledge-base/backend/.env

# 运行
./knowledge-base
```

**方式二：指定配置文件**

```bash
./knowledge-base -env /path/to/config/.env
```

**方式三：系统环境变量（推荐）**

```bash
# 在 systemd 服务中配置
# /etc/systemd/system/knowledge-base.service
[Service]
Environment="DB_HOST=127.0.0.1"
Environment="DB_PORT=3306"
Environment="DB_USER=kb_user"
Environment="DB_PASSWORD=YourPassword"
Environment="DB_NAME=knowledge_base"
Environment="JWT_SECRET=your-random-secret"
Environment="GIN_MODE=release"
```

**方式四：环境变量文件（systemd）**

```bash
# 创建环境变量文件
sudo tee /opt/knowledge-base/backend/.env << EOF
DB_HOST=127.0.0.1
DB_PASSWORD=YourPassword
JWT_SECRET=your-random-secret
GIN_MODE=release
EOF

# systemd 服务引用
[Service]
EnvironmentFile=/opt/knowledge-base/backend/.env
```

## 本地编译 + 上传部署

### 步骤 1：本地编译

```powershell
# Windows PowerShell
cd C:\Users\Administrator\.qclaw\workspace-agent-40338b03\knowledge-base\backend

# 设置交叉编译变量
$env:GOOS="linux"
$env:GOARCH="amd64"
$env:CGO_ENABLED="0"

# 编译（配置不打包进二进制）
go build -o knowledge-base main.go
```

### 步骤 2：上传文件

```powershell
# 上传二进制文件
scp knowledge-base user@server:/opt/knowledge-base/backend/

# 上传配置文件（首次部署或配置变更时）
scp .env.prod user@server:/opt/knowledge-base/backend/.env
```

### 步骤 3：服务器配置

```bash
# SSH 登录服务器
ssh user@server

# 创建配置文件（如果未上传）
cat > /opt/knowledge-base/backend/.env << 'EOF'
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=kb_user
DB_PASSWORD=YourStrongPassword123!
DB_NAME=knowledge_base
JWT_SECRET=your-random-jwt-secret-key-at-least-32-chars
GIN_MODE=release
EOF

# 设置权限
chmod 600 /opt/knowledge-base/backend/.env
chown nobody:nobody /opt/knowledge-base/backend/.env

# 重启服务
sudo systemctl restart knowledge-base
```

## 配置修改（无需重新编译）

编译后的二进制文件**不包含配置**，配置通过外部文件/环境变量提供。

### 修改数据库密码

```bash
# 1. 修改 .env 文件
vim /opt/knowledge-base/backend/.env
# 更新 DB_PASSWORD=NewPassword

# 2. 重启服务
sudo systemctl restart knowledge-base
```

### 切换数据库

```bash
# 修改 .env 文件
DB_HOST=新数据库地址
DB_PORT=新端口
DB_USER=新用户名
DB_PASSWORD=新密码

# 重启服务
sudo systemctl restart knowledge-base
```

## 配置项说明

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DB_HOST` | 数据库地址 | localhost |
| `DB_PORT` | 数据库端口 | 3306 |
| `DB_USER` | 数据库用户名 | root |
| `DB_PASSWORD` | 数据库密码 | 空 |
| `DB_NAME` | 数据库名 | knowledge_base |
| `JWT_SECRET` | JWT 签名密钥 | ⚠️ 必须修改 |
| `SERVER_PORT` | 服务端口 | 8080 |
| `GIN_MODE` | 运行模式 | debug |

## 安全建议

1. **生产环境必须修改**：
   - `JWT_SECRET`：至少 32 位随机字符串
   - `DB_PASSWORD`：强密码

2. **文件权限**：
   ```bash
   chmod 600 .env          # 仅所有者可读写
   chown nobody:nobody .env # 服务用户所有
   ```

3. **不要提交敏感配置**：
   - `.env` 已在 `.gitignore` 中
   - `.env.local` 已在 `.gitignore` 中
   - 仅提交 `.env.example` 作为模板

4. **生成随机密钥**：
   ```bash
   # Linux
   openssl rand -base64 32

   # 或
   head -c 32 /dev/urandom | base64
   ```

## .env.local 用途

`.env.local` 用于开发时的**本地个性化配置**，会覆盖 `.env` 中的值：

```bash
# 例如：不同开发者使用不同的数据库
# .env.local（不提交到 git）
DB_PORT=3308
DB_PASSWORD=我的个人密码
```

这样每个人可以有自己的本地配置，而不影响团队其他成员。
