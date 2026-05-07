package main

import (
	"log"

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

	// 注册路由
	routes.SetupRoutes(r)

	// 启动服务器
	addr := ":" + cfg.Server.Port
	log.Printf("服务器启动在 http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务器启动失败：%v", err)
	}
}