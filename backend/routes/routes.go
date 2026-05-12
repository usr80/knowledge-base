package routes

import (
	"knowledge-base/controllers"
	"knowledge-base/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// 初始化控制器
	authController := controllers.NewAuthController()
	userController := controllers.NewUserController()
	documentController := controllers.NewDocumentController()
	categoryController := controllers.NewCategoryController()
	chatController := controllers.NewChatController()
	searchController := controllers.NewSearchController()

	// 公开路由
	public := r.Group("/api")
	{
		// 认证
		public.POST("/auth/register", authController.Register)
		public.POST("/auth/login", authController.Login)
	}

	// 需要认证的路由
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		// 用户
		protected.GET("/user/profile", authController.GetProfile)
		protected.PUT("/user/profile", userController.UpdateProfile)
		protected.PUT("/user/password", userController.ChangePassword)

		// 文档
		protected.POST("/documents", documentController.Create)
		protected.GET("/documents", documentController.List)
		protected.GET("/documents/:id", documentController.GetByID)
		protected.PUT("/documents/:id", documentController.Update)
		protected.DELETE("/documents/:id", documentController.Delete)
		protected.POST("/documents/import", documentController.Import)
		protected.GET("/documents/:id/export/markdown", documentController.ExportMarkdown)

		// 分类
		protected.POST("/categories", categoryController.Create)
		protected.GET("/categories", categoryController.List)
		protected.GET("/categories/:id", categoryController.GetByID)
		protected.PUT("/categories/:id", categoryController.Update)
		protected.DELETE("/categories/:id", categoryController.Delete)

		// AI 问答
		protected.POST("/chat/ask", chatController.Ask)
		protected.POST("/chat/ask/stream", chatController.AskStream)
		protected.GET("/chat/sessions", chatController.ListSessions)
		protected.GET("/chat/sessions/:id", chatController.GetSession)
		protected.DELETE("/chat/sessions/:id", chatController.DeleteSession)

		// 用量统计
		protected.GET("/chat/usage/stats", chatController.GetUsageStats)
		protected.GET("/chat/usage/logs", chatController.GetUsageLogs)

		// 模型管理
		protected.GET("/models", chatController.ListModels)
		protected.POST("/models/select", chatController.SelectModel)

		// 文档索引
		protected.POST("/documents/:id/index", chatController.IndexDocument)

		// 搜索
		protected.GET("/search", searchController.Search)
		protected.POST("/search/rebuild", searchController.RebuildIndex)
	}
}