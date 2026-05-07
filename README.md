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
- **UI 组件**: Element Plus
- **状态管理**: Pinia
- **路由**: Vue Router
- **Markdown**: markdown-it

## 功能特性

### 第一阶段（MVP）✅
- [x] 用户注册/登录（JWT 认证）
- [x] 文档 CRUD（支持 Markdown）
- [x] 分类管理（支持多级分类）
- [x] 标签管理
- [x] 文档搜索（标题/内容）
- [x] 个人中心（资料修改、密码修改）

### 第二阶段（规划中）
- [ ] AI 智能问答（RAG 架构）
- [ ] 文档导入/导出（Markdown/PDF）
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

# 配置环境变量（可选，默认使用以下配置）
# export DB_HOST=localhost
# export DB_PORT=3307
# export DB_USER=root
# export DB_PASSWORD=your_password
# export DB_NAME=knowledge_base

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

## 项目结构

```
knowledge-base/
├── backend/
│   ├── config/          # 配置文件
│   ├── models/          # 数据模型
│   ├── controllers/     # 控制器
│   ├── services/        # 业务逻辑
│   ├── middleware/      # 中间件
│   ├── routes/          # 路由
│   └── main.go          # 入口文件
├── frontend/
│   ├── src/
│   │   ├── api/         # API 请求
│   │   ├── components/  # 组件
│   │   ├── views/       # 页面
│   │   ├── router/      # 路由
│   │   ├── stores/      # 状态管理
│   │   └── assets/      # 静态资源
│   └── package.json
└── docs/                # 文档
```

## API 接口

### 认证
- `POST /api/auth/register` - 用户注册
- `POST /api/auth/login` - 用户登录

### 用户
- `GET /api/user/profile` - 获取用户信息
- `PUT /api/user/profile` - 更新用户信息
- `PUT /api/user/password` - 修改密码

### 文档
- `GET /api/documents` - 文档列表
- `GET /api/documents/:id` - 获取文档详情
- `POST /api/documents` - 创建文档
- `PUT /api/documents/:id` - 更新文档
- `DELETE /api/documents/:id` - 删除文档

### 分类
- `GET /api/categories` - 分类列表
- `POST /api/categories` - 创建分类
- `PUT /api/categories/:id` - 更新分类
- `DELETE /api/categories/:id` - 删除分类

## 数据库配置

默认数据库配置（可通过环境变量覆盖）：

| 配置项 | 默认值 |
|--------|--------|
| Host | localhost |
| Port | 3307 |
| User | root |
| Password | j4PNMPGi52RAkDP2 |
| Database | knowledge_base |
| Charset | utf8mb4 |

## 开发者

「开发」- 资深程序员
技术栈：Go + Vue