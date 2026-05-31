# Lianji System Design - AI 建站系统

## 1. Overview

**Project Name:** Lianji (联极)  
**Domain:** lianji.win  
**Type:** AI-powered website builder (similar to v0.dev)  
**Core Functionality:** Users describe their website requirements → AI generates complete PHP website code → One-click deployment to Kangle virtual hosting

## 2. Technical Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.23 + Gin |
| Frontend | Go Templ + HTMX (TBD) |
| AI Model | MiniMax (via minimax-relay) |
| Virtual Hosting | Kangle EP |
| Database | Per-customer MySQL (via Kangle) |
| Storage | File system (JSON) |

## 3. Architecture

```
┌─────────────┐      ┌─────────────┐      ┌──────────────┐
│   User      │──────│  Lianji     │──────│   MiniMax    │
│   Browser   │      │  Backend    │      │   API        │
└─────────────┘      └─────────────┘      └──────────────┘
                            │
                            ▼
                     ┌──────────────┐
                     │   Kangle     │
                     │   EP Panel   │
                     │ (PHP + MySQL)│
                     └──────────────┘
```

## 4. Backend Structure

```
backend/
├── main.go              # Entry point, Gin HTTP server
├── go.mod               # Go modules
├── go.sum
├── models/
│   └── project.go       # Project, GenerateRequest, DeployRequest, DeployStatus
├── handlers/
│   └── handlers.go      # HTTP handlers for all API endpoints
├── kangle/
│   └── client.go        # Kangle EP API client (HTTP)
├── ai/
│   └── generator.go     # AI code generation module
└── storage/
    └── storage.go       # File-based storage (JSON)
```

## 5. API Endpoints

### Health Check
```
GET /health
Response: {"status": "ok", "service": "Lianji AI Builder", "version": "0.1.0", "kangle": "connected|disconnected"}
```

### Project Management
```
POST /api/v1/projects        # Create project
GET  /api/v1/projects        # List all projects
GET  /api/v1/projects/:id   # Get project details
DELETE /api/v1/projects/:id # Delete project
```

### AI Generation
```
POST /api/v1/generate
Body: {
  "prompt": "创建一个企业官网",
  "site_type": "企业站",
  "theme_color": "#007bff"
}
Response: {
  "success": true,
  "data": {
    "html_code": "...",
    "css_code": "..."
  }
}
```

### Deployment
```
POST /api/v1/deploy/:id
Body: {
  "domain": "example.lianji.win",
  "ftp_user": "user123",
  "ftp_password": "pass123"
}
Response: {"message": "部署已开始", "project_id": "...", "status": "deploying"}

GET /api/v1/deploy/status/:id
Response: {
  "project_id": "...",
  "status": "deployed",
  "progress": 100,
  "message": "部署完成",
  "site_url": "http://example.lianji.win"
}
```

## 6. Data Models

### Project
```go
type Project struct {
    ID          string    // UUID
    Name        string
    Description string
    Status      string    // pending, generating, generated, deploying, deployed, failed
    HTMLCode    string     // Generated HTML
    CSSCode     string     // Generated CSS
    PHPFiles    []string   // List of PHP files
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeployedURL string    // URL after deployment
    KangleVID   string    // Kangle virtual host ID
}
```

### GenerateRequest
```go
type GenerateRequest struct {
    Prompt     string // User's description
    SiteType   string // 企业站, 博客, 电商, 落地页, etc.
    ThemeColor string // Hex color code
}
```

### DeployRequest
```go
type DeployRequest struct {
    Domain      string
    SubDomain   string
    FTPHost     string
    FTPPort     int
    FTPUser     string
    FTPPassword string
}
```

### KangleVID
```go
type KangleVID struct {
    VID      string
    Host     string
    Username string
    Password string
    DBName   string
    DBUser   string
    DBPass   string
    FtpHost  string
    FtpPort  int
    FtpUser  string
    FtpPass  string
}
```

## 7. AI Generation Flow

1. User sends description via `/api/v1/generate`
2. Backend creates `GenerateRequest` with prompt, site_type, theme_color
3. AIGenerator calls MiniMax API via minimax-relay
4. MiniMax returns HTML + CSS code
5. Response sent back to user

### AI Prompt Template
```
你是一个专业的网站开发工程师。请根据用户的需求生成一个完整的网站。

要求：
- 生成现代化的、响应式的 HTML 和 CSS 代码
- 使用纯 HTML + CSS，不需要 JavaScript 框架
- 确保代码完整可以直接使用

用户需求：{prompt}

网站类型：{site_type}
主题颜色：{theme_color}

请生成：
1. 完整的 index.html 文件
2. 单独的 style.css 文件

只返回代码，不需要解释。
```

## 8. Deployment Flow

1. User calls `/api/v1/deploy/:id` with domain and FTP credentials
2. Backend creates virtual host on Kangle via API
3. Backend uploads generated files via FTP
4. Backend binds domain to the virtual host
5. User can access the deployed website

### Kangle API Integration
- Protocol: HTTP API (JSON)
- Default port: 7777
- Authentication: API Key in request token field

## 9. Configuration (Environment Variables)

| Variable | Description | Example |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `KANGLE_HOST` | Kangle server IP | `192.168.1.100` |
| `KANGLE_API_KEY` | Kangle API key | `xxx` |
| `AI_API_HOST` | MiniMax relay URL | `http://localhost:3000` |
| `AI_API_KEY` | MiniMax API key | `sk-xxx` |

## 10. Project Status

- [x] Backend core structure
- [x] Project management API
- [x] AI generation API (stub)
- [ ] System design document
- [ ] Kangle integration (needs credentials)
- [ ] Frontend development
- [ ] FTP upload implementation
- [ ] Database integration
- [ ] User authentication
- [ ] Production deployment

## 11. Open Questions

1. Should we support multiple PHP files or just single-page sites initially?
2. Do we need user accounts, or is this a single-tenant system?
3. Should we store Kangle credentials per-project or globally?
4. How to handle custom domain binding (CNAME vs reverse proxy)?
