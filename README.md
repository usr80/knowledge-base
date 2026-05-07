# 个人知识库系统

基于 Go + Vue 3 的多用户个人知识库系统，支持 Markdown 文档管理、分类标签、全文搜索，预留 AI 问答（RAG）架构。

## 技术栈

### 后端
- **语言**: Go 1.21+
- **框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL 5.7+
- **认证**: JWT

### 前端
- **框架**: Vue 3 + Vite
- **UI 组件**: Vuetify 3 (Material Design)
- **状态管理**: Pinia
- **路由**: Vue Router
- **图标**: @mdi/font

## 功能特性

### 第一阶段（MVP）✅
- [x] 用户注册/登录（JWT 认证）
- [x] 文档 CRUD（支持 Markdown）
- [x] 分类管理（支持多级分类）
- [x] 标签管理
- [x] 文档搜索（标题/内容）
- [x] 个人中心（资料修改、密码修改）
- [x] Material Design UI

### 第二阶段（进行中）
- [ ] AI 智能问答（RAG 架构）
- [x] 文档导入/导出（Markdown/PDF）
- [ ] 全文搜索引擎（Elasticsearch）
- [ ] 文档版本历史
- [ ] 协作分享功能

## 快速开始

### 环境要求
- Go 1.21+
- Node.js 18+
- MySQL 5.7+

### 后端启动

```bash
cd backend

# 下载依赖
go mod tidy

# 复制配置文件
cp .env.example .env

# 编辑配置（开发环境示例）
vim .env.dev

# 运行开发环境
go run main.go -e dev

# 运行
go run main.go
```

后端服务默认运行在 `http://localhost:8080`

### 前端启动

```bash
cd frontend

# 安装依赖
npm install

# 开发模式
npm run dev

# 构建生产版本
npm run build
```

前端服务默认运行在 `http://localhost:3000`

## 构建部署

### 分离构建脚本

```bash
# Windows - 仅编译后端
powershell -File deploy/build-backend.ps1

# Windows - 仅构建前端
powershell -File deploy/build-frontend.ps1

# Linux - 仅编译后端
./deploy/build-backend.sh

# Linux - 仅构建前端
./deploy/build-frontend.sh
```

### 完整构建（前后端一起）

```bash
# Windows
powershell -File deploy/build.ps1

# Linux
./deploy/build.sh
```

构建输出：
- 后端：`build/backend/knowledge-base` (Linux amd64 ELF)
- 前端：`build/frontend/dist/` (静态资源)

### Docker 部署

```bash
# 构建并启动
docker-compose up -d

# 查看日志
docker logs -f knowledge-base

# 停止
docker-compose down
```

### CentOS 部署

详见 [docs/DEPLOY_CENTOS.md](docs/DEPLOY_CENTOS.md)

## 配置管理

支持多环境配置，优先级：命令行参数 > 环境变量 > .env.{env} > .env > 默认值

```bash
# 开发环境（默认）
./knowledge-base

# 指定环境
./knowledge-base -e dev
./knowledge-base -e test
./knowledge-base -e prod

# 指定配置文件
./knowledge-base -env /path/to/.env

# 调试配置
./knowledge-base -show-config
```

详见 [docs/CONFIG.md](docs/CONFIG.md)

## 项目结构

```
knowledge-base/
├── backend/
│   ├── config/          # 配置管理
│   ├── models/          # 数据模型
│   ├── controllers/     # 控制器
│   ├── services/        # 业务逻辑
│   ├── middleware/      # 中间件
│   ├── routes/          # 路由
│   ├── .env.example     # 配置模板
│   └── main.go          # 入口文件
├── frontend/
│   ├── src/
│   │   ├── api/         # API 请求封装
│   │   ├── views/       # 页面组件
│   │   ├── router/      # 路由配置
│   │   ├── stores/      # Pinia 状态
│   │   └── main.js      # 入口文件
│   ├── vite.config.js   # Vite 配置
│   └── package.json
├── deploy/              # 部署脚本
│   ├── build.ps1        # Windows 完整构建
│   ├── build.sh         # Linux 完整构建
│   ├── build-backend.ps1
│   ├── build-backend.sh
│   ├── build-frontend.ps1
│   └── build-frontend.sh
├── docs/                # 文档
│   ├── DEPLOY_CENTOS.md # CentOS 部署指南
│   ├── CONFIG.md        # 配置管理指南
│   └── DEV_LOG.md       # 开发日志
├── docker-compose.yml   # Docker 编排
├── Dockerfile           # Docker 镜像
└── README.md
```

## API 接口

### 认证
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/auth/register | 用户注册 |
| POST | /api/auth/login | 用户登录 |

### 用户
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/user/profile | 获取用户信息 |
| PUT | /api/user/profile | 更新用户信息 |
| PUT | /api/user/password | 修改密码 |

### 文档
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/documents | 文档列表（分页、搜索） |
| GET | /api/documents/:id | 获取文档详情 |
| POST | /api/documents | 创建文档 |
| PUT | /api/documents/:id | 更新文档 |
| DELETE | /api/documents/:id | 删除文档 |
| POST | /api/documents/import | 导入 Markdown 文件 |
| GET | /api/documents/:id/export/markdown | 导出 Markdown 文件 |

### 分类
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/categories | 分类列表 |
| POST | /api/categories | 创建分类 |
| PUT | /api/categories/:id | 更新分类 |
| DELETE | /api/categories/:id | 删除分类 |

## 数据库配置

| 配置项 | 环境变量 | 默认值 |
|--------|----------|--------|
| 主机 | DB_HOST | localhost |
| 端口 | DB_PORT | 3306 |
| 用户 | DB_USER | root |
| 密码 | DB_PASSWORD | (空) |
| 数据库 | DB_NAME | knowledge_base |

数据表：
- `users` - 用户表
- `categories` - 分类表
- `tags` - 标签表
- `documents` - 文档表
- `document_tags` - 文档标签关联表

## 开发者

「开发」- 资深程序员  
技术栈：Go + Vue  
风格：代码优雅美观、逻辑清晰
