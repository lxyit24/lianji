package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AIGenerator AI 网站代码生成器
type AIGenerator struct {
	apiHost string
	apiKey  string
	client  *http.Client
}

// NewAIGenerator 创建 AI 生成器
func NewAIGenerator(apiHost, apiKey string) *AIGenerator {
	if !strings.HasPrefix(apiHost, "http") {
		apiHost = "https://" + apiHost
	}
	
	return &AIGenerator{
		apiHost: apiHost,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 120 * time.Second, // AI 生成可能需要较长时间
		},
	}
}

// GenerateRequest 生成请求
type GenerateRequest struct {
	Prompt     string `json:"prompt"`
	SiteType   string `json:"site_type"`   // 企业站、博客、电商、落地页
	ThemeColor string `json:"theme_color"` // 主题色
}

// GenerateResponse 生成响应
type GenerateResponse struct {
	HTMLCode string   `json:"html_code"` // HTML 代码
	CSSCode  string   `json:"css_code"`  // CSS 代码
	PHPFiles []PHPFile `json:"php_files"` // PHP 文件
}

// PHPFile PHP 文件
type PHPFile struct {
	Path    string `json:"path"`    // 文件路径
	Content string `json:"content"` // 文件内容
}

// Generate 使用 AI 生成网站代码
func (g *AIGenerator) Generate(req GenerateRequest) (*GenerateResponse, error) {
	// 构建 Prompt
	systemPrompt := `你是一个专业的网站开发工程师。用户需要你根据描述生成一个完整的网站。

请生成一个现代、美观的网站，包含：

1. **HTML 结构**：语义化、SEO 友好
2. **CSS 样式**：响应式设计，现代化 UI，支持深色模式
3. **PHP 后端**（如需要）：轻量、安全

网站类型：` + req.SiteType + `

要求：
- 使用纯 PHP（无框架依赖，直接可部署到标准 PHP 虚拟主机）
- 数据库使用 MySQL，连接信息通过环境变量或配置文件获取
- 代码简洁、安全（防 SQL 注入、XSS）
- 包含必要的错误处理

输出格式（严格遵循 JSON）：
{
  "html_code": "<!DOCTYPE html>...",
  "css_code": "body { ... }",
  "php_files": [
    {"path": "index.php", "content": "<?php ... ?>"},
    {"path": "config.php", "content": "<?php ... ?>"}
  ]
}

只输出 JSON，不要输出任何其他内容。`

	userPrompt := "请为以下需求生成网站代码：" + req.Prompt

	if req.ThemeColor != "" {
		userPrompt += "\n\n主题色偏好：" + req.ThemeColor
	}

	// 构建请求体
	requestBody := map[string]interface{}{
		"model": "MiniMax-M2.7-highspeed",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.7,
		"max_tokens": 4096,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %v", err)
	}

	// 发送请求
	url := g.apiHost + "/v1/chat/completions"
	if !strings.Contains(g.apiHost, "/v1") {
		url = g.apiHost + "/v1/chat/completions"
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("AI 请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI 返回错误: %s - %s", resp.Status, string(body))
	}

	// 解析响应
	var aiResp map[string]interface{}
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return nil, fmt.Errorf("解析 AI 响应失败: %v", err)
	}

	// 提取 AI 回复内容
	choices, ok := aiResp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("AI 响应格式错误")
	}

	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("AI 响应 choices 格式错误")
	}

	message, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("AI 响应 message 格式错误")
	}

	content, ok := message["content"].(string)
	if !ok {
		return nil, fmt.Errorf("AI 响应 content 格式错误")
	}

	// 尝试提取 JSON
	result, err := g.parseAIResponse(content)
	if err != nil {
		// 如果解析失败，尝试修复 JSON
		return g.parseAIResponse(g.fixJSON(content))
	}

	return result, nil
}

// parseAIResponse 解析 AI 返回的 JSON 内容
func (g *AIGenerator) parseAIResponse(content string) (*GenerateResponse, error) {
	// 尝试提取 JSON 块
	content = strings.TrimSpace(content)
	
	// 移除 markdown 代码块标记
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
	}
	if strings.HasSuffix(content, "```") {
		content = strings.TrimSuffix(content, "```")
	}
	
	content = strings.TrimSpace(content)

	var result GenerateResponse
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %v", err)
	}

	return &result, nil
}

// fixJSON 尝试修复不完整的 JSON
func (g *AIGenerator) fixJSON(content string) string {
	// 尝试找到 JSON 开始和结束位置
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	
	if start == -1 || end == -1 || end < start {
		return content
	}
	
	return content[start : end+1]
}
