package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"docs-hub/internal/config"
)

// ServiceDoc 服务文档
type ServiceDoc struct {
	Name        string                 `json:"name"`
	DisplayName string                 `json:"display_name"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Owner       string                 `json:"owner"`
	Tags        []string               `json:"tags"`
	BaseURL     string                 `json:"base_url"`
	DocURL      string                 `json:"doc_url"`
	Healthy     bool                   `json:"healthy"`
	LastUpdated time.Time              `json:"last_updated"`
	SwaggerDoc  map[string]interface{} `json:"swagger_doc,omitempty"`
}

// DocService 文档服务
type DocService struct {
	config   *config.Config
	docs     map[string]*ServiceDoc
	mu       sync.RWMutex
	client   *http.Client
}

// NewDocService 创建文档服务
func NewDocService(cfg *config.Config) *DocService {
	ds := &DocService{
		config: cfg,
		docs:   make(map[string]*ServiceDoc),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// 初始化所有服务
	for _, svc := range cfg.Services {
		ds.docs[svc.Name] = &ServiceDoc{
			Name:        svc.Name,
			DisplayName: svc.DisplayName,
			Description: svc.Description,
			Status:      svc.Status,
			Owner:       svc.Owner,
			Tags:        svc.Tags,
			BaseURL:     svc.BaseURL,
			DocURL:      svc.BaseURL + svc.DocPath,
			Healthy:     false,
		}
	}

	// 初始加载文档
	ds.RefreshAll()

	return ds
}

// StartRefreshLoop 启动刷新循环
func (ds *DocService) StartRefreshLoop() {
	ticker := time.NewTicker(ds.config.GetRefreshDuration())
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Refreshing all service docs...")
		ds.RefreshAll()
	}
}

// RefreshAll 刷新所有服务文档
func (ds *DocService) RefreshAll() {
	var wg sync.WaitGroup
	for _, svc := range ds.config.Services {
		wg.Add(1)
		go func(svc config.ServiceConfig) {
			defer wg.Done()
			ds.RefreshService(svc.Name)
		}(svc)
	}
	wg.Wait()
}

// RefreshService 刷新单个服务文档
func (ds *DocService) RefreshService(name string) error {
	ds.mu.Lock()
	doc, exists := ds.docs[name]
	ds.mu.Unlock()

	if !exists {
		return fmt.Errorf("service %s not found", name)
	}

	// 获取服务配置
	var svcConfig *config.ServiceConfig
	for _, svc := range ds.config.Services {
		if svc.Name == name {
			svcConfig = &svc
			break
		}
	}

	if svcConfig == nil {
		return fmt.Errorf("service config %s not found", name)
	}

	// 健康检查
	healthy := ds.checkHealth(svcConfig.BaseURL + svcConfig.HealthCheck)

	// 获取文档
	swaggerDoc, err := ds.fetchSwaggerDoc(svcConfig.BaseURL + svcConfig.DocPath)

	ds.mu.Lock()
	defer ds.mu.Unlock()

	doc.Healthy = healthy
	doc.LastUpdated = time.Now()

	if err == nil && swaggerDoc != nil {
		doc.SwaggerDoc = swaggerDoc
		log.Printf("Successfully refreshed docs for %s", name)
	} else {
		log.Printf("Failed to refresh docs for %s: %v", name, err)
	}

	return err
}

// checkHealth 检查服务健康状态
func (ds *DocService) checkHealth(url string) bool {
	resp, err := ds.client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// fetchSwaggerDoc 获取 Swagger 文档
func (ds *DocService) fetchSwaggerDoc(url string) (map[string]interface{}, error) {
	resp, err := ds.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch doc: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var doc map[string]interface{}
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, err
	}

	return doc, nil
}

// GetCatalog 获取服务目录
func (ds *DocService) GetCatalog() []*ServiceDoc {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var catalog []*ServiceDoc
	for _, doc := range ds.docs {
		// 返回不包含完整文档的摘要
		catalog = append(catalog, &ServiceDoc{
			Name:        doc.Name,
			DisplayName: doc.DisplayName,
			Description: doc.Description,
			Status:      doc.Status,
			Owner:       doc.Owner,
			Tags:        doc.Tags,
			BaseURL:     doc.BaseURL,
			DocURL:      doc.DocURL,
			Healthy:     doc.Healthy,
			LastUpdated: doc.LastUpdated,
		})
	}

	return catalog
}

// GetServiceDoc 获取服务文档
func (ds *DocService) GetServiceDoc(name string) (*ServiceDoc, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	doc, exists := ds.docs[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}

	return doc, nil
}

// GetSwaggerJSON 获取 Swagger JSON
func (ds *DocService) GetSwaggerJSON(name string) (map[string]interface{}, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	doc, exists := ds.docs[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}

	if doc.SwaggerDoc == nil {
		return nil, fmt.Errorf("swagger doc not available for %s", name)
	}

	return doc.SwaggerDoc, nil
}
