# Pear 项目管理系统 (Go重写版)

> 基于 pearProjectApi (PHP) 和 pearProject (Vue) 重写的 Go 版本项目管理系统

## 📁 项目结构

```
ms_project/
├── backend/                      # 后端服务
│   ├── api/                     # HTTP API 服务
│   │   ├── api/                 # API 处理器（按模块组织）
│   │   ├── internal/            # 内部实现
│   │   ├── pkg/                 # 公共包
│   │   ├── config/              # 配置文件
│   │   └── main.go              # 入口文件
│   ├── user/                    # 用户 gRPC 服务
│   ├── project/                 # 项目 gRPC 服务
│   ├── common/                  # 公共库
│   ├── grpc/                    # gRPC 协议定义
│   ├── config/                  # 全局配置
│   ├── go.work                  # Go workspace 配置
│   └── docker-compose.yaml      # Docker 编排
│
├── frontend/                     # 前端服务 (Vue.js)
│   ├── src/                     # 源代码
│   │   ├── api/                 # API接口定义
│   │   ├── components/          # Vue组件
│   │   ├── views/               # 页面视图
│   │   ├── router/              # 路由配置
│   │   ├── store/               # Vuex状态管理
│   │   └── assets/              # 静态资源
│   ├── public/                  # 静态资源
│   └── package.json             # 依赖管理
│
├── database/                     # 数据库相关
│   ├── msproject_full_dump.sql  # 完整数据库导出
│   ├── msproject_clean.sql      # 清理后的数据库
│   ├── msproject_import.sql     # 数据导入
│   ├── create_missing_tables.sql # 创建缺失表
│   ├── insert_notify.sql        # 插入通知数据
│   └── template_fix.sql         # 模板修复
│
├── scripts/                      # 部署和构建脚本
│   ├── build.bat                # 构建脚本
│   ├── run.bat                  # 运行脚本
│   ├── export_database.ps1      # 数据库导出
│   └── import_database.ps1      # 数据库导入
│
├── docs/                         # 项目文档
│   ├── README.md                # 项目说明
│   └── PROJECT_STRUCTURE.md     # 结构说明
│
├── .gitignore                   # Git 忽略配置
└── .git/                        # Git 仓库
```

## 🚀 快速开始

### 环境要求

- Go 1.18+
- Node.js 14+
- MySQL 5.7+
- Redis 6.0+

### 后端启动

```bash
# 进入后端目录
cd backend

# 启动 API 服务
cd api && go run main.go

# 启动用户服务
cd user && go run main.go

# 启动项目服务
cd project && go run main.go
```

### 前端启动

```bash
# 进入前端目录
cd frontend

# 安装依赖
npm install

# 开发模式
npm run serve

# 生产构建
npm run build
```

## 🏗️ 架构说明

### 微服务架构

项目采用微服务架构，包含以下服务：

1. **api** - HTTP API 网关，提供 RESTful 接口
2. **user** - 用户管理服务（gRPC）
3. **project** - 项目管理服务（gRPC）
4. **common** - 公共工具库

### 功能模块

#### 用户管理模块
- ✅ 用户登录/注册
- ✅ 个人信息管理
- ✅ 密码修改
- ✅ 头像上传
- ✅ 组织切换

#### 组织管理模块
- ✅ 创建组织
- ✅ 组织列表
- ✅ 编辑组织
- ✅ 删除组织
- ✅ 退出组织

#### 项目管理模块
- ✅ 项目CRUD
- ✅ 项目成员管理
- ✅ 项目收藏
- ✅ 项目模板
- ✅ 项目版本管理
- ✅ 项目功能特性

#### 任务管理模块
- ✅ 任务CRUD
- ✅ 任务成员管理
- ✅ 任务标签
- ✅ 任务阶段
- ✅ 任务工作流
- ✅ 任务工时记录

#### 其他模块
- ✅ 部门管理
- ✅ 邀请链接
- ✅ 文件管理
- ✅ 通知系统
- ✅ 权限控制

## 📊 技术栈

### 后端

- **语言**: Go 1.18
- **框架**: Gin (HTTP), gRPC
- **数据库**: MySQL 5.7+, Redis 6.0+
- **ORM**: GORM
- **认证**: JWT
- **服务发现**: Etcd

### 前端

- **框架**: Vue.js 2.x
- **UI**: Ant Design Vue
- **状态管理**: Vuex
- **路由**: Vue Router
- **HTTP**: Axios

## 🔄 从 PHP 版本迁移

本项目从 `pearProjectApi` (PHP) 重写为 Go 版本，主要变更：

1. **后端语言**: PHP → Go
2. **架构模式**: 单体应用 → 微服务架构
3. **性能提升**: 更好的并发处理能力
4. **代码组织**: 更清晰的模块划分

## 📝 API模块列表

### 核心模块
- `account` - 账户管理
- `auth` - 认证授权
- `user` - 用户管理
- `index` - 系统配置

### 组织项目模块
- `organization` - 组织管理
- `project` - 项目管理
- `project_member` - 项目成员
- `project_collect` - 项目收藏
- `project_template` - 项目模板
- `project_version` - 项目版本
- `project_features` - 项目特性
- `project_info` - 项目信息

### 任务模块
- `task` - 任务管理
- `task_member` - 任务成员
- `task_tag` - 任务标签
- `task_stages` - 任务阶段
- `task_stages_template` - 任务阶段模板
- `task_workflow` - 任务工作流

### 其他模块
- `department` - 部门管理
- `department_member` - 部门成员
- `invite_link` - 邀请链接
- `file` - 文件管理
- `notify` - 通知管理
- `menu` - 菜单管理
- `node` - 节点管理
- `events` - 事件管理
- `source_link` - 资源链接

## 🐛 问题反馈

如有问题，请提交 Issue 或联系维护团队。

## 📄 License

MIT License
