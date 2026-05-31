package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"lianji/backend/models"
)

// Storage 存储接口
type Storage interface {
	SaveProject(project *models.Project) error
	GetProject(id string) (*models.Project, error)
	ListProjects() ([]*models.Project, error)
	DeleteProject(id string) error
}

// FileStorage 基于文件的存储
type FileStorage struct {
	basePath string
	mu       sync.RWMutex
}

// NewFileStorage 创建文件存储
func NewFileStorage(basePath string) *FileStorage {
	// 确保目录存在
	os.MkdirAll(basePath, 0755)
	os.MkdirAll(filepath.Join(basePath, "projects"), 0755)
	
	return &FileStorage{
		basePath: basePath,
	}
}

// SaveProject 保存项目
func (s *FileStorage) SaveProject(project *models.Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 设置时间
	now := time.Now()
	if project.CreatedAt.IsZero() {
		project.CreatedAt = now
	}
	project.UpdatedAt = now
	
	// 生成文件路径
	filePath := filepath.Join(s.basePath, "projects", project.ID+".json")
	
	// 序列化
	data, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化项目失败: %v", err)
	}
	
	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}
	
	return nil
}

// GetProject 获取项目
func (s *FileStorage) GetProject(id string) (*models.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	filePath := filepath.Join(s.basePath, "projects", id+".json")
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("项目不存在: %s", id)
		}
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}
	
	var project models.Project
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("解析项目失败: %v", err)
	}
	
	return &project, nil
}

// ListProjects 列出所有项目
func (s *FileStorage) ListProjects() ([]*models.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	projectsDir := filepath.Join(s.basePath, "projects")
	
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*models.Project{}, nil
		}
		return nil, fmt.Errorf("读取目录失败: %v", err)
	}
	
	projects := make([]*models.Project, 0, len(entries))
	
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		
		id := entry.Name()[:len(entry.Name())-5] // 去掉 .json
		project, err := s.GetProject(id)
		if err != nil {
			continue // 跳过无效文件
		}
		projects = append(projects, project)
	}
	
	return projects, nil
}

// DeleteProject 删除项目
func (s *FileStorage) DeleteProject(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	filePath := filepath.Join(s.basePath, "projects", id+".json")
	
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("项目不存在: %s", id)
		}
		return fmt.Errorf("删除文件失败: %v", err)
	}
	
	return nil
}
