package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/ronaldocristover/lms-backend/docs"
	"github.com/ronaldocristover/lms-backend/internal/config"
	"github.com/ronaldocristover/lms-backend/internal/middleware"
)

func setupRoutes(router *gin.Engine, h *handlers, cfg *config.Config) {
	router.Use(middleware.Recovery(nil))
	router.Use(middleware.Logger(nil))
	router.Use(middleware.CORS(middleware.CORSConfig{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           int(cfg.CORS.MaxAge.Seconds()),
	}))
	router.Use(middleware.RequestID())
	router.Use(middleware.RateLimit())

	registerHealthRoutes(router, h)
	registerSwaggerRoutes(router)
	registerAuthRoutes(router, h)
	registerProtectedRoutes(router, h, cfg)
	registerUploadRoutes(router, h, cfg)
}

func registerHealthRoutes(router *gin.Engine, h *handlers) {
	router.GET("/health", h.Health.HealthCheck)
	router.GET("/health/status", h.Health.HealthStatusCheck)
	router.GET("/health/detailed", h.Health.HealthDetailedCheck)
	router.GET("/health/live", h.Health.HealthLiveCheck)
	router.GET("/health/ready", h.Health.HealthReadyCheck)
}

func registerSwaggerRoutes(router *gin.Engine) {
	//nolint:staticcheck // gin-swagger uses deprecated FileServer
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func registerAuthRoutes(router *gin.Engine, h *handlers) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.RefreshToken)
	}
}

func registerProtectedRoutes(router *gin.Engine, h *handlers, cfg *config.Config) {
	protected := router.Group("")
	protected.Use(middleware.Auth(cfg.JWT.Secret))
	{
		protected.GET("/auth/me", h.Auth.Me)

		roles := protected.Group("/roles")
		{
			roles.POST("", h.Role.Create)
			roles.GET("", h.Role.List)
			roles.GET("/:id", h.Role.Get)
			roles.PUT("/:id", h.Role.Update)
			roles.DELETE("/:id", h.Role.Delete)
		}

		users := protected.Group("/users")
		{
			users.POST("", h.User.Create)
			users.GET("", h.User.List)
			users.GET("/:id", h.User.Get)
			users.PUT("/:id", h.User.Update)
			users.DELETE("/:id", h.User.Delete)
		}

		organizations := protected.Group("/organizations")
		{
			organizations.GET("", h.Org.List)
			organizations.GET("/:id", h.Org.Get)
			organizations.POST("", h.Org.Create)
			organizations.PUT("/:id", h.Org.Update)
			organizations.DELETE("/:id", h.Org.Delete)
			organizations.GET("/:id/users", h.Org.ListUsers)
			organizations.POST("/:id/users", h.Org.AddUser)
			organizations.PUT("/:id/users/:userId", h.Org.UpdateUserRole)
			organizations.DELETE("/:id/users/:userId", h.Org.RemoveUser)
		}

		languages := protected.Group("/languages")
		{
			languages.POST("", h.Language.Create)
			languages.GET("", h.Language.List)
			languages.GET("/:id", h.Language.Get)
			languages.PUT("/:id", h.Language.Update)
			languages.DELETE("/:id", h.Language.Delete)
		}

		medias := protected.Group("/media")
		{
			medias.POST("", h.Media.Create)
			medias.GET("", h.Media.List)
			medias.GET("/:id", h.Media.Get)
			medias.PUT("/:id", h.Media.Update)
			medias.DELETE("/:id", h.Media.Delete)
		}

		subtitles := protected.Group("/subtitles")
		{
			subtitles.POST("", h.Subtitle.Create)
			subtitles.GET("", h.Subtitle.List)
			subtitles.GET("/:id", h.Subtitle.Get)
			subtitles.PUT("/:id", h.Subtitle.Update)
			subtitles.DELETE("/:id", h.Subtitle.Delete)
		}

		categories := protected.Group("/categories")
		{
			categories.POST("", h.Category.Create)
			categories.GET("", h.Category.List)
			categories.GET("/:id", h.Category.Get)
			categories.PUT("/:id", h.Category.Update)
			categories.DELETE("/:id", h.Category.Delete)
		}

		series := protected.Group("/series")
		{
			series.POST("", h.Series.Create)
			series.GET("", h.Series.List)
			series.GET("/:id", h.Series.Get)
			series.PUT("/:id", h.Series.Update)
			series.DELETE("/:id", h.Series.Delete)
		}

		sessions := protected.Group("/sessions")
		{
			sessions.POST("", h.Session.Create)
			sessions.GET("", h.Session.List)
			sessions.GET("/:id", h.Session.Get)
			sessions.PUT("/:id", h.Session.Update)
			sessions.DELETE("/:id", h.Session.Delete)
		}

		contents := protected.Group("/contents")
		{
			contents.POST("", h.Content.Create)
			contents.GET("", h.Content.List)
			contents.GET("/:id", h.Content.Get)
			contents.PUT("/:id", h.Content.Update)
			contents.DELETE("/:id", h.Content.Delete)
		}
	}
}

func registerUploadRoutes(router *gin.Engine, h *handlers, cfg *config.Config) {
	router.POST("/upload", middleware.Auth(cfg.JWT.Secret), h.Upload.Upload)
	router.GET("/uploads/:filename", h.Upload.Serve)
}
