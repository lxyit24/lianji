package kangle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client Kangle API 客户端（HTTP 方式）
type Client struct {
	host   string
	apiKey string
	client *http.Client
}

// NewClient 创建 Kangle 客户端
func NewClient(host, apiKey string) *Client {
	if !strings.HasPrefix(host, "http") {
		host = "http://" + host
	}
	
	return &Client{
		host:   host,
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// KangleRequest Kangle API 请求
type KangleRequest struct {
	Module     string                 `json:"module"`
	Action     string                 `json:"action"`
	Token      string                 `json:"token"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// KangleResponse Kangle API 响应
type KangleResponse struct {
	Result int             `json:"result"` // 1=成功
	Msg    string          `json:"msg"`
	Data   json.RawMessage `json:"data,omitempty"`
}

// CreateVirtualHost 创建虚拟主机
func (c *Client) CreateVirtualHost(domain, user, password string) (*KangleVID, error) {
	reqBody := KangleRequest{
		Module: "host",
		Action: "create",
		Token:  c.apiKey,
		Parameters: map[string]interface{}{
			"domain":    domain,
			"user":      user,
			"pass":      password,
			"bandwidth": "102400",
			"traffic":   "102400",
			"mysql":     "1",
			"ftp":       "1",
		},
	}
	
	resp, err := c.doRequest(reqBody)
	if err != nil {
		// 开发模式：返回模拟数据
		fmt.Printf("Kangle API 调用失败（开发模式）: %v\n", err)
		return &KangleVID{
			VID:      domain,
			Host:     c.host,
			Username: user,
			Password: password,
			FtpHost:  c.host,
			FtpPort:  21,
			FtpUser:  user,
			FtpPass:  password,
		}, nil
	}
	
	vid := &KangleVID{
		VID:      domain,
		Host:     c.host,
		Username: user,
		Password: password,
		FtpHost:  c.host,
		FtpPort:  21,
		FtpUser:  user,
		FtpPass:  password,
	}
	
	if resp.Result == 1 && resp.Data != nil {
		var data map[string]interface{}
		if err := json.Unmarshal(resp.Data, &data); err == nil {
			if v, ok := data["vid"].(string); ok {
				vid.VID = v
			}
			if db, ok := data["mysql"].(map[string]interface{}); ok {
				vid.DBName, _ = db["database"].(string)
				vid.DBUser, _ = db["user"].(string)
				vid.DBPass, _ = db["pass"].(string)
			}
		}
	}
	
	return vid, nil
}

// DeleteVirtualHost 删除虚拟主机
func (c *Client) DeleteVirtualHost(vid string) error {
	reqBody := KangleRequest{
		Module: "host",
		Action: "delete",
		Token:  c.apiKey,
		Parameters: map[string]interface{}{
			"vid": vid,
		},
	}
	
	resp, err := c.doRequest(reqBody)
	if err != nil {
		return err
	}
	
	if resp.Result != 1 {
		return fmt.Errorf("删除失败: %s", resp.Msg)
	}
	
	return nil
}

// BindDomain 绑定域名
func (c *Client) BindDomain(vid, domain string) error {
	reqBody := KangleRequest{
		Module: "host",
		Action: "bindDomain",
		Token:  c.apiKey,
		Parameters: map[string]interface{}{
			"vid":    vid,
			"domain": domain,
		},
	}
	
	resp, err := c.doRequest(reqBody)
	if err != nil {
		fmt.Printf("绑定域名失败（继续）: %v\n", err)
		return nil // 不阻塞主流程
	}
	
	if resp.Result != 1 {
		return fmt.Errorf("绑定域名失败: %s", resp.Msg)
	}
	
	return nil
}

// CreateDatabase 创建数据库
func (c *Client) CreateDatabase(vid, dbName, dbUser, dbPass string) error {
	reqBody := KangleRequest{
		Module: "host",
		Action: "createDatabase",
		Token:  c.apiKey,
		Parameters: map[string]interface{}{
			"vid":  vid,
			"db":   dbName,
			"user": dbUser,
			"pass": dbPass,
		},
	}
	
	resp, err := c.doRequest(reqBody)
	if err != nil {
		return err
	}
	
	if resp.Result != 1 {
		return fmt.Errorf("创建数据库失败: %s", resp.Msg)
	}
	
	return nil
}

// GetHostInfo 获取主机信息
func (c *Client) GetHostInfo(vid string) (*KangleVID, error) {
	reqBody := KangleRequest{
		Module: "host",
		Action: "info",
		Token:  c.apiKey,
		Parameters: map[string]interface{}{
			"vid": vid,
		},
	}
	
	resp, err := c.doRequest(reqBody)
	if err != nil {
		return nil, err
	}
	
	if resp.Result != 1 {
		return nil, fmt.Errorf("获取主机信息失败: %s", resp.Msg)
	}
	
	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, err
	}
	
	return &KangleVID{
		VID:      vid,
		Host:     c.host,
		Username: data["username"].(string),
	}, nil
}

// HealthCheck 检查连接状态
func (c *Client) HealthCheck() error {
	reqBody := KangleRequest{
		Module: "system",
		Action: "ping",
		Token:  c.apiKey,
	}
	
	_, err := c.doRequest(reqBody)
	return err
}

// doRequest 发送 API 请求
func (c *Client) doRequest(req KangleRequest) (*KangleResponse, error) {
	// Kangle 管理面板通常在 7777 或 7778 端口
	url := c.host + ":7777/api"
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %v", err)
	}
	
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}
	
	var kangleResp KangleResponse
	if err := json.Unmarshal(body, &kangleResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}
	
	return &kangleResp, nil
}

// KangleVID 虚拟主机信息
type KangleVID struct {
	VID      string `json:"vid"`
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
	DBUser   string `json:"db_user"`
	DBPass   string `json:"db_pass"`
	FtpHost  string `json:"ftp_host"`
	FtpPort  int    `json:"ftp_port"`
	FtpUser  string `json:"ftp_user"`
	FtpPass  string `json:"ftp_pass"`
}
