# Pear 项目管理系统

> 基于 Go + Vue.js 的现代化项目管理系统

## 🎯 项目简介

本项目是从 PHP 版本 (pearProjectApi) 重写为 Go 版本的项目管理系统，采用微服务架构，提供完整的项目管理、任务管理、团队协作等功能。

## 📁 项目结构

```
ms_project/
├── backend/          # 后端服务 (Go)
├── frontend/         # 前端服务 (Vue.js)
├── database/         # 数据库文件
├── scripts/          # 部署脚本
└── docs/             # 项目文档
```

详细结构请查看 [PROJECT_STRUCTURE.md](docs/PROJECT_STRUCTURE.md)

## 🚀 快速开始

### 后端启动

```bash
cd backend/api && go run main.go
```

### 前端启动

```bash
cd frontend && npm install && npm run serve
```

## 📖 文档

- [项目说明](docs/README.md)
- [目录结构](docs/PROJECT_STRUCTURE.md)

## 🛠️ 技术栈

**后端:** Go 1.18 + Gin + gRPC + MySQL + Redis

**前端:** Vue.js 2.x + Ant Design Vue + Vuex

## 📄 License

MIT License
