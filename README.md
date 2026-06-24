# Minlink - 轻量级短链服务系统

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Minlink 是一个基于 Go + Gin + SQLite 的轻量级短链服务系统，采用纯本地架构，无需 Redis、MySQL 等外部依赖，部署简单，响应极速。

## 核心特性

### 🚀 极速响应
- 短链跳转毫秒级完成（< 10ms）
- 纯 SQLite 本地查询，无网络 IO 延迟
- 本地内存计数聚合，减少数据库写入压力

### 📦 零依赖部署
- 无需 Redis、MySQL 等外部服务
- 单一可执行文件，资源占用极低
- Docker 一键部署

### 🔐 安全可靠
- JWT 无状态认证
- BCrypt 密码加密
- 管理员权限控制

### 🎯 功能完善
- 短链生成与管理
- 自定义过期时间
- 访问量统计
- 用户管理（管理员）

## 抸术栈

| 组件 | 技术 | 说明 |
|------|------|------|
| Web 框架 | Gin | 高性能轻量级框架 |
| 数据库 | SQLite | 嵌入式数据库，零配置 |
| ORM | GORM | 功能完善的 ORM 库 |
| ID 生成 | Snowflake | 本地生成，全局唯一 |
| 认证 | JWT + bcrypt | 无状态认证，密码加密 |
| 前端 | HTML5 + CSS3 + JS | 原生技术栈，零依赖 |

## 快速开始

### 本地运行

```bash
# 克隆项目
git clone https://github.com/yourusername/minlink.git
cd minlink

# 安装依赖
go mod download

# 运行服务
go run ./cmd/main.go
```

访问 http://localhost:8080 即可使用。

### Docker 部署

```bash
# 构建并启动（使用默认版本号）
docker-compose up -d --build

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### Docker 镜像版本管理

```bash
# 构建指定版本的镜像
APP_VERSION=1.0.0 docker-compose build

# 使用指定版本启动
APP_VERSION=1.0.0 docker-compose up -d

# 推送镜像到私有仓库（需要配置 DOCKER_REGISTRY）
DOCKER_REGISTRY=registry.example.com/ APP_VERSION=1.0.0 docker-compose push

# 只使用已构建的镜像（不重新构建）
docker-compose up -d --no-build
```

**环境变量配置（.env）：**

| 变量 | 说明 | 默认值 |
|------|------|--------|
| APP_VERSION | 镜像版本号 | 1.0.0 |
| DOCKER_REGISTRY | 镜像仓库地址（可选） | 空 |
| APP_ENV | 运行环境 | production |
| PORT | 服务端口 | 8080 |

## 核心功能

### 1. 短链生成

- 输入长链接，自动生成短码
- 支持设置短链名称和备注
- 支持自定义过期时间（1天/3天/1周/1个月/1年/长期）
- 自动使用当前域名生成短链

### 2. 短链管理

- 查看所有短链列表
- 显示短链名称、备注、访问量、剩余有效天数
- 支持复制、禁用、删除操作
- 搜索过滤功能

### 3. 用户管理（管理员）

- 用户列表查看
- 创建新用户账号
- 编辑用户信息（用户名、邮箱、角色、状态）
- 禁用/启用用户
- 删除用户

### 4. 访问统计

- 实时访问量统计
- 本地内存聚合计数
- 定时批量刷盘

## API 接口

### 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/:shortCode` | 短链重定向（302） |
| POST | `/api/v1/short-url` | 创建短链 |
| GET | `/api/v1/short-url/list` | 获取短链列表 |

### 用户接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/login` | 用户登录 |
| GET | `/api/v1/user/profile` | 获取用户信息 |
| PUT | `/api/v1/user/password` | 修改密码 |

### 管理员接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/admin/users` | 创建用户 |
| GET | `/api/v1/admin/users` | 用户列表 |
| PUT | `/api/v1/admin/users/:id` | 更新用户 |
| DELETE | `/api/v1/admin/users/:id` | 删除用户 |
| GET | `/api/v1/admin/short-url/list` | 短链列表（管理） |
| PUT | `/api/v1/admin/short-url/:code/status` | 更新短链状态 |
| DELETE | `/api/v1/admin/short-url/:code` | 删除短链 |

## 默认账号

系统启动后自动创建管理员账号：

- **用户名**: `admin`
- **密码**: `admin123`

> ⚠️ 请在生产环境中及时修改默认密码

## 项目结构

```
minlink/
├── cmd/main.go              # 主入口
├── internal/
│   ├── handler/             # HTTP 处理器
│   ├── service/             # 业务逻辑
│   ├── repository/          # 数据访问
│   ├── model/               # 数据模型
│   ├── middleware/          # 中间件
│   └── util/                # 工具函数
├── static/                  # 前端资源
│   ├── index.html
│   ├── css/main.css
│   └── js/app.js
├── Dockerfile
├── docker-compose.yml
└── .env                     # 环境配置
```

## 环境配置

```env
APP_ENV=production
PORT=8080
DB_PATH=./data/minlink.db
JWT_SECRET=your-secret-key
```

## 性能指标

| 指标 | 数值 |
|------|------|
| 短链跳转响应 | < 10ms |
| SQLite 查询 | < 1ms |
| 内存占用 | ~20MB |
| 单文件部署 | ~15MB |

## 扩展方向

- 数据量增长：可迁移至 PostgreSQL/MySQL
- 高并发场景：可选引入 Redis 缓存
- 分布式部署：Snowflake ID 支持多节点
- 自定义域名：支持多域名绑定
- 统计分析：可扩展访问日志分析

## License

MIT License

---

**Minlink** - 轻量、极速、零依赖的短链服务系统