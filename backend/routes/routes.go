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

		// 分类
		protected.POST("/categories", categoryController.Create)
		protected.GET("/categories", categoryController.List)
		protected.GET("/categories/:id", categoryController.GetByID)
		protected.PUT("/categories/:id", categoryController.Update)
		protected.DELETE("/categories/:id", categoryController.Delete)
	}
}