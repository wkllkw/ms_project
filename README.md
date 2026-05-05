# Pear 项目管理系统

> 基于 Go + Vue.js 的现代化项目管理系统

## 🎯 项目简介

本项目是从 PHP 版本 (pearProjectApi) 重写为 Go 版本的项目管理系统，采用微服务架构，提供完整的项目管理、任务管理、团队协作等功能。

## 📁 项目结构

```
ms_project/
├── backend/          # 后端服务 (Go)
│   ├── api/          # API 网关 (Gin, 端口 18000)
│   ├── user/         # 用户微服务 (gRPC, 端口 18881)
│   ├── project/      # 项目微服务 (gRPC, 端口 18882)
│   ├── common/       # 公共库 (discovery, encrypts, logs 等)
│   └── docker-compose.yaml  # 基础设施容器 (MySQL, Redis, etcd)
├── frontend/         # 前端服务 (Vue.js, 端口 8045)
├── deploy/           # 部署脚本
│   └── docker-helper.sh  # Docker 容器管理辅助脚本
├── data/             # Docker 数据卷
└── docs/             # 项目文档
```

## 🚀 快速开始

### 环境要求

- Go 1.18+
- Node.js 16+ (推荐使用 nvm 管理)
- Docker & Docker Compose
- MySQL 8.0 / Redis 6 / etcd v3.5 (由 Docker Compose 提供)

### 一键启动（systemctl）

项目已配置 systemd 服务，支持一键启动和管理：

```bash
# 启动所有服务
sudo systemctl start ms-project-user.service
sudo systemctl start ms-project-project.service
sudo systemctl start ms-project-api.service
sudo systemctl start ms-project-frontend.service

# 查看所有服务状态
systemctl status ms-project-{user,project,api,frontend}.service

# 停止所有服务
sudo systemctl stop ms-project-frontend.service
sudo systemctl stop ms-project-api.service
sudo systemctl stop ms-project-project.service
sudo systemctl stop ms-project-user.service

# 开机自启
sudo systemctl enable ms-project-{user,project,api,frontend}.service

# 取消开机自启
sudo systemctl disable ms-project-{user,project,api,frontend}.service

# 查看服务日志
journalctl -u ms-project-api.service -f        # 实时跟踪 API 网关日志
journalctl -u ms-project-user.service -f       # 用户服务日志
journalctl -u ms-project-project.service -f    # 项目服务日志
journalctl -u ms-project-frontend.service -f   # 前端日志
```

### 服务说明

| 服务名 | 二进制文件 | 端口 | 说明 |
|--------|-----------|------|------|
| `ms-project-user` | `backend/user/user_bin` | 18881 (gRPC) | 用户认证、验证码、注册 |
| `ms-project-project` | `backend/project/project_bin` | 18882 (gRPC) | 项目/任务/文件管理 |
| `ms-project-api` | `backend/api/project-api` | 18000 (HTTP) | API 网关，转发至 gRPC |
| `ms-project-frontend` | `npm run serve` | 8045 (HTTP) | Vue.js 前端 |

> **启动顺序**: Docker 容器 → user → project → api → frontend  
> **停止顺序**: frontend → api → project → user

### Docker 基础设施

MySQL、Redis、etcd 由 Docker Compose 管理，systemd 服务会在启动前自动检查并启动容器：

```bash
# 手动管理 Docker 容器
cd backend && docker-compose up -d       # 启动
cd backend && docker-compose stop        # 停止（保留数据）
cd backend && docker-compose down        # 停止并删除容器

# 也可使用辅助脚本
deploy/docker-helper.sh ensure          # 确保容器运行
deploy/docker-helper.sh stop-if-idle    # 若后端服务全部停止则停止容器
```

| 容器 | 端口映射 | 说明 |
|------|---------|------|
| mysql8 | 13306→3306 | MySQL 8.0 |
| redis6 | 6379→6379 | Redis 6 |
| etcd3 | 2379→2379 | etcd v3.5 (服务注册发现) |

### 手动启动（开发调试）

```bash
# 1. 启动 Docker 基础设施
cd backend && docker-compose up -d

# 2. 启动后端微服务（按顺序）
cd backend/user && go run .
cd backend/project && go run .
cd backend/api && go run .

# 3. 启动前端
cd frontend && npm run serve
```

### 编译更新二进制文件

修改代码后，需重新编译并重启对应服务：

```bash
# 编译单个服务
cd backend/api && go build -o project-api .
cd backend/user && go build -o user_bin .
cd backend/project && go build -o project_bin .

# 重启服务
sudo systemctl restart ms-project-api.service
sudo systemctl restart ms-project-user.service
sudo systemctl restart ms-project-project.service
```

### 默认账号

| 账号 | 密码 | 说明 |
|------|------|------|
| admin | 123456 | 管理员 |
| rookie | 123456 | 普通用户 |
| zhangmy | 123456 | 普通用户 |

> 密码存储方式：前端 MD5 哈希 → 后端再 MD5 → 存入数据库（双重 MD5）。生产环境请及时修改默认密码。

### 配置文件

| 服务 | 配置文件 |
|------|---------|
| API 网关 | `backend/api/config/config.yaml` |
| 用户服务 | `backend/user/config/config.yaml` |
| 项目服务 | `backend/project/config/config.yaml` |
| Docker | `backend/.env` + `backend/docker-compose.yaml` |

关键配置项：
- `rpc.userAddr` / `rpc.projectAddr`：gRPC 直连地址（配置后绕过 etcd 发现）
- `etcd.addrs`：etcd 集群地址（不配置 rpc 直连时使用）
- `mysql` / `redis`：数据库和缓存连接信息
- API 网关支持环境变量覆盖邮件配置：`MAIL_HOST`, `MAIL_PORT`, `MAIL_USER`, `MAIL_PASSWORD`, `MAIL_FROM`

## 📖 文档

- [项目说明](docs/README.md)
- [目录结构](docs/PROJECT_STRUCTURE.md)

## 🛠️ 技术栈

**后端:** Go 1.18 + Gin + gRPC + MySQL + Redis + etcd

**前端:** Vue.js 2.x + Ant Design Vue + Vuex

## 📄 License

MIT License
