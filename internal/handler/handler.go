package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"docs-hub/internal/service"
)

// Response 通用响应
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "docs-hub",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// GetServices 获取服务列表（供前端静态页面使用）
func GetServices(ds *service.DocService) gin.HandlerFunc {
	return func(c *gin.Context) {
		catalog := ds.GetCatalog()
		// 转换为前端需要的格式
		services := make([]map[string]interface{}, 0, len(catalog))
		for _, doc := range catalog {
			services = append(services, map[string]interface{}{
				"name":        doc.Name,
				"title":       doc.DisplayName,
				"description": doc.Description,
				"status":      doc.Status,
				"healthy":     doc.Healthy,
				"owner":       doc.Owner,
				"tags":        doc.Tags,
				"swaggerUrl":  "/api/docs/" + doc.Name + "/swagger.json",
			})
		}
		c.JSON(http.StatusOK, services)
	}
}

// GetCatalog 获取服务目录
func GetCatalog(ds *service.DocService) gin.HandlerFunc {
	return func(c *gin.Context) {
		catalog := ds.GetCatalog()
		c.JSON(http.StatusOK, Response{
			Code:    0,
			Message: "success",
			Data:    catalog,
		})
	}
}

// GetServiceDoc 获取服务文档
func GetServiceDoc(ds *service.DocService) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceName := c.Param("service")
		swagger, err := ds.GetSwaggerJSON(serviceName)
		if err != nil {
			c.JSON(http.StatusNotFound, Response{
				Code:    404,
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, swagger)
	}
}

// GetServiceInfo 获取服务详情
func GetServiceInfo(ds *service.DocService) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceName := c.Param("service")
		doc, err := ds.GetServiceDoc(serviceName)
		if err != nil {
			c.JSON(http.StatusNotFound, Response{
				Code:    404,
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, Response{
			Code:    0,
			Message: "success",
			Data:    doc,
		})
	}
}

// RefreshDocs 刷新所有文档
func RefreshDocs(ds *service.DocService) gin.HandlerFunc {
	return func(c *gin.Context) {
		go ds.RefreshAll()
		c.JSON(http.StatusOK, Response{
			Code:    0,
			Message: "Refresh started",
		})
	}
}

// RefreshServiceDoc 刷新单个服务文档
func RefreshServiceDoc(ds *service.DocService) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceName := c.Param("service")
		err := ds.RefreshService(serviceName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Code:    500,
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, Response{
			Code:    0,
			Message: "Refresh completed",
		})
	}
}

// IndexPage 首页
func IndexPage(ds *service.DocService) gin.HandlerFunc {
	return func(c *gin.Context) {
		catalog := ds.GetCatalog()
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":    "API Documentation Portal",
			"services": catalog,
		})
	}
}

// ServiceDocsPage 服务文档页面
func ServiceDocsPage(ds *service.DocService) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceName := c.Param("service")
		doc, err := ds.GetServiceDoc(serviceName)
		if err != nil {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"message": err.Error(),
			})
			return
		}
		c.HTML(http.StatusOK, "swagger.html", gin.H{
			"title":       doc.DisplayName + " API",
			"service":     doc,
			"swaggerURL":  "/api/docs/" + serviceName + "/swagger.json",
		})
	}
}
