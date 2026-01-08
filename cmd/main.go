package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"docs-hub/internal/config"
	"docs-hub/internal/handler"
	"docs-hub/internal/service"
)

func main() {
	// 加载配置
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/dev.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建文档服务
	docService := service.NewDocService(cfg)

	// 启动后台刷新任务
	go docService.StartRefreshLoop()

	// 设置 Gin 模式
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)

	// 创建路由
	r := gin.Default()

	// 健康检查
	r.GET("/health", handler.HealthCheck)

	// API 路由
	api := r.Group("/api")
	{
		// 获取服务列表（供静态页面使用）
		api.GET("/services", handler.GetServices(docService))

		// 获取服务目录
		api.GET("/catalog", handler.GetCatalog(docService))

		// 获取服务文档
		api.GET("/docs/:service/swagger.json", handler.GetServiceDoc(docService))

		// 获取服务详情
		api.GET("/services/:service", handler.GetServiceInfo(docService))

		// 刷新文档
		api.POST("/refresh", handler.RefreshDocs(docService))
		api.POST("/refresh/:service", handler.RefreshServiceDoc(docService))
	}

	// 静态文件（统一使用 web/static）
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")

	// 首页（使用静态文件）
	r.StaticFile("/", "./web/static/index.html")

	// 服务文档页面（保留模板渲染作为备用）
	r.GET("/docs/:service", handler.ServiceDocsPage(docService))

	// 启动服务
	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Server.Port
	}
	addr := ":" + port

	log.Printf("Docs Hub starting on %s", addr)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Services configured: %d", len(cfg.Services))

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
