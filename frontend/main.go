package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"lianji/frontend/pages"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// 静态文件服务
	router.Static("/static", "./static")

	// 主页
	router.GET("/", func(c *gin.Context) {
		var buf bytes.Buffer
		err := pages.Index().Render(context.Background(), &buf)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
	})

	// 代理后端 API
	router.POST("/api/v1/generate", proxyToBackend)
	router.GET("/api/v1/projects", proxyToBackend)
	router.POST("/api/v1/projects", proxyToBackend)
	router.GET("/api/v1/projects/:id", proxyToBackend)
	router.DELETE("/api/v1/projects/:id", proxyToBackend)
	router.POST("/api/v1/deploy/:id", proxyToBackend)
	router.GET("/api/v1/deploy/status/:id", proxyToBackend)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "Lianji Frontend",
			"version": "0.1.0",
		})
	})

	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8082"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("Frontend running on :%s\n", port)
	fmt.Printf("Backend URL: %s\n", backendURL)

	router.Run(":" + port)
}

func proxyToBackend(c *gin.Context) {
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8082"
	}

	// 构建目标 URL
	targetURL := backendURL + c.Request.RequestURI

	// 创建新请求
	req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to create request"})
		return
	}

	// 复制请求头
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Backend connection failed"})
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 复制响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to read response"})
		return
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}
