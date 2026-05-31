package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"lianji/backend/ai"
	"lianji/backend/kangle"
	"lianji/backend/models"
	"lianji/backend/storage"
)

// Handler 处理器
type Handler struct {
	store        storage.Storage
	kangleClient *kangle.Client
	aiGenerator  *ai.AIGenerator
}

// NewHandler 创建处理器
func NewHandler(store storage.Storage, kangleClient *kangle.Client, aiGenerator *ai.AIGenerator) *Handler {
	return &Handler{
		store:        store,
		kangleClient: kangleClient,
		aiGenerator:  aiGenerator,
	}
}

// HealthCheck 健康检查
func (h *Handler) HealthCheck(c *gin.Context) {
	status := map[string]interface{}{
		"status":    "ok",
		"service":   "Lianji AI Builder",
		"timestamp": time.Now().Unix(),
		"version":   "0.1.0",
	}

	// 检查 Kangle 连接
	if h.kangleClient != nil {
		if err := h.kangleClient.HealthCheck(); err != nil {
			status["kangle"] = "disconnected"
		} else {
			status["kangle"] = "connected"
		}
	}

	c.JSON(http.StatusOK, status)
}

// CreateProject 创建项目
func (h *Handler) CreateProject(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project := &models.Project{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.store.SaveProject(project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// ListProjects 列出项目
func (h *Handler) ListProjects(c *gin.Context) {
	projects, err := h.store.ListProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"total":    len(projects),
	})
}

// GetProject 获取项目
func (h *Handler) GetProject(c *gin.Context) {
	id := c.Param("id")

	project, err := h.store.GetProject(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject 删除项目
func (h *Handler) DeleteProject(c *gin.Context) {
	id := c.Param("id")

	if err := h.store.DeleteProject(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "项目已删除"})
}

// GenerateSite 生成网站
func (h *Handler) GenerateSite(c *gin.Context) {
	var req models.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取项目 ID（可选）
	projectID := c.Query("project_id")

	// 设置默认值
	if req.SiteType == "" {
		req.SiteType = "企业站"
	}

	// 获取 AI 生成器
	generator := h.aiGenerator
	if generator == nil {
		generator = ai.NewAIGenerator(
			os.Getenv("AI_API_HOST"),
			os.Getenv("AI_API_KEY"),
		)
	}

	// 调用 AI 生成
	result, err := generator.Generate(ai.GenerateRequest{
		Prompt:     req.Prompt,
		SiteType:   req.SiteType,
		ThemeColor: req.ThemeColor,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("AI 生成失败: %v", err)})
		return
	}

	// 如果有项目 ID，更新项目
	if projectID != "" {
		project, err := h.store.GetProject(projectID)
		if err == nil {
			project.Status = "generated"
			project.HTMLCode = result.HTMLCode
			project.CSSCode = result.CSSCode
			project.UpdatedAt = time.Now()
			h.store.SaveProject(project)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// DeployToKangle 部署到 Kangle
func (h *Handler) DeployToKangle(c *gin.Context) {
	id := c.Param("id")

	var req models.DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取项目
	project, err := h.store.GetProject(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}

	// 更新状态
	project.Status = "deploying"
	project.UpdatedAt = time.Now()
	h.store.SaveProject(project)

	// 在 goroutine 中执行部署
	go func() {
		h.deployProject(project, req)
	}()

	c.JSON(http.StatusOK, gin.H{
		"message":   "部署已开始",
		"project_id": project.ID,
		"status":    "deploying",
	})
}

// deployProject 执行部署
func (h *Handler) deployProject(project *models.Project, req models.DeployRequest) {
	// 1. 创建虚拟主机
	kangleClient := kangle.NewClient(
		os.Getenv("KANGLE_HOST"),
		os.Getenv("KANGLE_API_KEY"),
	)

	vid, err := kangleClient.CreateVirtualHost(req.Domain, req.FTPUser, req.FTPPassword)
	if err != nil {
		project.Status = "failed"
		project.UpdatedAt = time.Now()
		h.store.SaveProject(project)
		return
	}

	project.KangleVID = vid.VID

	// 2. 上传文件（通过 FTP）
	if err := h.uploadFilesFTP(project, req); err != nil {
		project.Status = "failed"
		project.UpdatedAt = time.Now()
		h.store.SaveProject(project)
		return
	}

	// 3. 绑定域名
	if err := kangleClient.BindDomain(vid.VID, req.Domain); err != nil {
		fmt.Printf("绑定域名失败: %v\n", err)
	}

	// 4. 完成
	project.Status = "deployed"
	project.DeployedURL = "http://" + req.Domain
	project.UpdatedAt = time.Now()
	h.store.SaveProject(project)
}

// uploadFilesFTP 通过 FTP 上传文件
func (h *Handler) uploadFilesFTP(project *models.Project, req models.DeployRequest) error {
	// 创建临时目录存放生成的文件
	tmpDir := filepath.Join(os.TempDir(), "lianji", project.ID)
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	// 写入 HTML 文件
	if project.HTMLCode != "" {
		os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte(project.HTMLCode), 0644)
	}

	// 写入 CSS 文件
	if project.CSSCode != "" {
		os.WriteFile(filepath.Join(tmpDir, "style.css"), []byte(project.CSSCode), 0644)
	}

	// TODO: 实现 FTP 上传
	// 这里需要使用 goftp 库上传文件

	return nil
}

// GetDeployStatus 获取部署状态
func (h *Handler) GetDeployStatus(c *gin.Context) {
	id := c.Param("id")

	project, err := h.store.GetProject(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}

	status := &models.DeployStatus{
		ProjectID: project.ID,
		Status:    project.Status,
		SiteURL:   project.DeployedURL,
	}

	switch project.Status {
	case "pending":
		status.Progress = 0
		status.Message = "等待开始"
	case "generating":
		status.Progress = 30
		status.Message = "AI 正在生成网站..."
	case "generated":
		status.Progress = 50
		status.Message = "网站生成完成，等待部署"
	case "deploying":
		status.Progress = 75
		status.Message = "正在部署到服务器..."
	case "deployed":
		status.Progress = 100
		status.Message = "部署完成"
	case "failed":
		status.Progress = 0
		status.Message = "部署失败"
	}

	c.JSON(http.StatusOK, status)
}

// UploadFile 上传文件（用于前端实时预览）
func (h *Handler) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的文件"})
		return
	}
	defer file.Close()

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}

	// 保存到临时目录
	tmpDir := filepath.Join(os.TempDir(), "lianji", "uploads")
	os.MkdirAll(tmpDir, 0755)

	filePath := filepath.Join(tmpDir, header.Filename)
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"filename":    header.Filename,
		"size":        header.Size,
		"preview_url": "/preview/" + header.Filename,
	})
}
