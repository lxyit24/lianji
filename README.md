# Lianji 快速启动指南

## 系统状态 ✅

| 组件 | 端口 | 状态 | 进程 |
|------|------|------|------|
| 后端 | 8082 | ✅ 运行中 | lianji-bin (PID 818893) |
| 前端 | 3001 | ✅ 运行中 | lianji-frontend (PID 830290) |

## 访问方式

### 本地访问
```bash
# 前端 UI
curl http://[::1]:3001/

# 后端 API
curl http://[::1]:8082/health
```

### 健康检查
```bash
# 后端
curl http://[::1]:8082/health
# {"status":"ok","service":"Lianji AI Builder","version":"0.1.0","kangle":"disconnected"}

# 前端
curl http://[::1]:3001/health
# {"service":"Lianji Frontend","status":"ok","version":"0.1.0"}
```

## API 端点

### 项目管理
- `GET /api/v1/projects` - 列出所有项目
- `POST /api/v1/projects` - 创建新项目
- `GET /api/v1/projects/:id` - 获取项目详情
- `DELETE /api/v1/projects/:id` - 删除项目

### AI 生成
- `POST /api/v1/generate` - AI 生成网站代码

### 部署
- `POST /api/v1/deploy/:id` - 部署到 Kangle
- `GET /api/v1/deploy/status/:id` - 获取部署状态

## 项目结构

```
lianji/
├── backend/
│   ├── main.go
│   ├── handlers/
│   ├── models/
│   ├── ai/
│   ├── kangle/
│   ├── storage/
│   ├── lianji-bin (编译产物)
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── main.go
│   ├── pages/
│   │   └── index.templ
│   ├── lianji-frontend (编译产物)
│   ├── go.mod
│   └── go.sum
├── docs/
│   └── SYSTEM_DESIGN.md
└── .gitignore
```

## 重启服务

### 后端
```bash
cd /root/.openclaw/workspace/lianji/backend
PORT=8082 ./lianji-bin &
```

### 前端
```bash
cd /root/.openclaw/workspace/lianji/frontend
PORT=3001 ./lianji-frontend &
```

## 下一步

1. **获取 Kangle 凭证** - 配置服务器连接
2. **数据库集成** - 添加持久化存储
3. **用户认证** - 实现登录系统
4. **GitHub Push** - 在其他网络环境推送代码

## 已知问题

- ⚠️ GitHub push 失败（网络问题）
- ⚠️ Kangle 未连接（需要凭证）
- ℹ️ 服务仅在 IPv6 上监听（需要用 `[::1]` 访问）
