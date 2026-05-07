package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"knowledge-base/config"
	"knowledge-base/middleware"
	"knowledge-base/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()
	log.Printf("配置加载成功，服务器端口：%s", cfg.Server.Port)

	// 初始化数据库
	if err := config.InitDB(&cfg.Database); err != nil {
		log.Fatalf("数据库初始化失败：%v", err)
	}

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 创建 Gin 引擎
	r := gin.Default()

	// 注册中间件
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.LoggerMiddleware())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 注册 API 路由
	routes.SetupRoutes(r)

	// 托管前端静态文件（Docker 部署时使用）
	staticDir := getEnv("STATIC_DIR", "")
	if staticDir != "" {
		setupSPA(r, staticDir)
		log.Printf("前端静态文件目录：%s", staticDir)
	}

	// 启动服务器
	addr := ":" + cfg.Server.Port
	log.Printf("服务器启动在 http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务器启动失败：%v", err)
	}
}

// setupSPA 配置 SPA 静态文件托管
// /assets 等静态资源直接返回文件
// 其他非 /api 请求返回 index.html（交给前端路由处理）
func setupSPA(r *gin.Engine, staticDir string) {
	// 静态资源（JS/CSS/图片/字体等）
	r.Static("/assets", filepath.Join(staticDir, "assets"))
	r.StaticFile("/favicon.ico", filepath.Join(staticDir, "favicon.ico"))
	r.StaticFile("/vite.svg", filepath.Join(staticDir, "vite.svg"))

	// SPA 路由：所有非 /api、非静态文件的请求返回 index.html
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// API 请求返回 404
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "接口不存在"})
			return
		}
		// 静态资源请求返回 404（避免把 .js/.css 当作 SPA 路由）
		if isStaticFile(path) {
			c.Status(http.StatusNotFound)
			return
		}
		// 其他请求返回 index.html
		c.File(filepath.Join(staticDir, "index.html"))
	})
}

// isStaticFile 判断是否为静态资源请求
func isStaticFile(path string) bool {
	ext := filepath.Ext(path)
	if ext == "" {
		return false
	}
	staticExts := map[string]bool{
		".js": true, ".css": true, ".png": true, ".jpg": true,
		".jpeg": true, ".gif": true, ".svg": true, ".ico": true,
		".woff": true, ".woff2": true, ".ttf": true, ".eot": true,
		".map": true, ".json": true, ".xml": true, ".txt": true,
	}
	return staticExts[ext]
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}