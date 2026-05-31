package main

import (
	"log"
	"os"

	"lianji/backend/ai"
	"lianji/backend/handlers"
	"lianji/backend/kangle"
	"lianji/backend/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化存储
	store := storage.NewFileStorage("./data")

	// 初始化 Kangle 客户端
	kangleClient := kangle.NewClient(
		os.Getenv("KANGLE_HOST"),
		os.Getenv("KANGLE_API_KEY"),
	)

	// 初始化 AI 生成器
	aiGenerator := ai.NewAIGenerator(
		os.Getenv("AI_API_HOST"),
		os.Getenv("AI_API_KEY"),
	)

	// 初始化处理器
	h := handlers.NewHandler(store, kangleClient, aiGenerator)

	// 启动 Gin 服务
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 健康检查
	r.GET("/health", h.HealthCheck)

	// API 路由
	api := r.Group("/api/v1")
	{
		// 项目管理
		api.POST("/projects", h.CreateProject)
		api.GET("/projects", h.ListProjects)
		api.GET("/projects/:id", h.GetProject)
		api.DELETE("/projects/:id", h.DeleteProject)

		// AI 生成
		api.POST("/generate", h.GenerateSite)

		// 部署到 Kangle
		api.POST("/deploy/:id", h.DeployToKangle)
		api.GET("/deploy/status/:id", h.GetDeployStatus)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf(" Lianji 服务启动，监听端口: %s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
