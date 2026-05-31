package models

import (
	"time"
)

// Project 网站项目
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // pending, generating, generated, deploying, deployed, failed
	HTMLCode    string    `json:"html_code,omitempty"`
	CSSCode     string    `json:"css_code,omitempty"`
	PHPFiles    []string  `json:"php_files,omitempty"` // PHP 文件列表
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeployedURL string    `json:"deployed_url,omitempty"` // 部署后的访问 URL
	KangleVID   string    `json:"kangle_vid,omitempty"`   // Kangle 虚拟主机 ID
}

// GenerateRequest AI 生成请求
type GenerateRequest struct {
	Prompt     string `json:"prompt" binding:"required"`
	SiteType   string `json:"site_type"` // 企业站、博客、电商、落地页等
	ThemeColor string `json:"theme_color"`
}

// DeployRequest 部署请求
type DeployRequest struct {
	Domain      string `json:"domain" binding:"required"`
	SubDomain   string `json:"sub_domain"`
	FTPHost     string `json:"ftp_host"`
	FTPPort     int    `json:"ftp_port"`
	FTPUser     string `json:"ftp_user"`
	FTPPassword string `json:"ftp_password"`
}

// DeployStatus 部署状态
type DeployStatus struct {
	ProjectID string `json:"project_id"`
	Status    string `json:"status"` // uploading, creating_db, binding_domain, done, failed
	Progress  int    `json:"progress"` // 0-100
	Message   string `json:"message"`
	SiteURL   string `json:"site_url,omitempty"`
}
