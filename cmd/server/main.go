// @title           LMS Backend API
// @version         1.0.0
// @description     Learning Management System Backend API
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  admin@lms.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /
// @schemes   http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/yourusername/lms/docs"
	"github.com/yourusername/lms/internal/config"
	"github.com/yourusername/lms/internal/handler"
	"github.com/yourusername/lms/internal/middleware"
	"github.com/yourusername/lms/internal/model"
	"github.com/yourusername/lms/internal/repository"
	"github.com/yourusername/lms/internal/scheduler"
	"github.com/yourusername/lms/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var sugar *zap.SugaredLogger
	if cfg.Server.Env == "production" {
		prodLog, _ := zap.NewProduction()
		sugar = prodLog.Sugar()
	} else {
		devLog, _ := zap.NewDevelopment()
		sugar = devLog.Sugar()
	}
	defer sugar.Sync()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
	)

	gormConfig := &gorm.Config{}
	if cfg.Server.Env != "production" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		sugar.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		sugar.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpen)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdle)

	if err := db.AutoMigrate(&model.Role{}, &model.User{}, &model.Organization{}, &model.OrganizationUser{}); err != nil {
		sugar.Fatalf("Failed to migrate database: %v", err)
	}

	roleRepo := repository.NewRoleRepository(db)
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo, roleRepo, cfg.JWT.Secret)
	roleSvc := service.NewRoleService(roleRepo)
	authHandler := handler.NewAuthHandler(userSvc)
	userHandler := handler.NewUserHandler(userSvc)
	roleHandler := handler.NewRoleHandler(roleSvc)

	orgRepo := repository.NewOrganizationRepository(db)
	orgUserRepo := repository.NewOrganizationUserRepository(db)
	orgSvc := service.NewOrganizationService(orgRepo, orgUserRepo, userRepo)
	orgHandler := handler.NewOrganizationHandler(orgSvc)

	sched := scheduler.NewScheduler(cfg.Jobs.Workers, cfg.Jobs.QueueSize)
	sched.Start()
	defer sched.Stop()

	version := "1.0.0"
	healthHandler := handler.NewHealthHandler(db, version)

	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	router.Use(middleware.Recovery(sugar))
	router.Use(middleware.Logger(sugar))
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.RateLimit())

	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/health/status", healthHandler.HealthStatusCheck)
	router.GET("/health/detailed", healthHandler.HealthDetailedCheck)
	router.GET("/health/live", healthHandler.HealthLiveCheck)
	router.GET("/health/ready", healthHandler.HealthReadyCheck)

	// Swagger documentation
	//nolint:staticcheck // gin-swagger uses deprecated FileServer
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	protected := router.Group("")
	protected.Use(middleware.Auth(cfg.JWT.Secret))
	{
		protected.GET("/auth/me", authHandler.Me)

		roles := protected.Group("/roles")
		{
			roles.POST("", roleHandler.Create)
			roles.GET("", roleHandler.List)
			roles.GET("/:id", roleHandler.Get)
			roles.PUT("/:id", roleHandler.Update)
			roles.DELETE("/:id", roleHandler.Delete)
		}

		users := protected.Group("/users")
		{
			users.POST("", userHandler.Create)
			users.GET("", userHandler.List)
			users.GET("/:id", userHandler.Get)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}

		organizations := protected.Group("/organizations")
		{
			organizations.GET("", orgHandler.List)
			organizations.GET("/:id", orgHandler.Get)
			organizations.POST("", orgHandler.Create)
			organizations.PUT("/:id", orgHandler.Update)
			organizations.DELETE("/:id", orgHandler.Delete)
			organizations.GET("/:id/users", orgHandler.ListUsers)
			organizations.POST("/:id/users", orgHandler.AddUser)
			organizations.PUT("/:id/users/:userId", orgHandler.UpdateUserRole)
			organizations.DELETE("/:id/users/:userId", orgHandler.RemoveUser)
		}
	}

	uploadHandler := handler.NewUploadHandler(cfg.Upload.Dir, cfg.Upload.MaxSize)
	router.POST("/upload", middleware.Auth(cfg.JWT.Secret), uploadHandler.Upload)
	router.GET("/uploads/:filename", uploadHandler.Serve)

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	go func() {
		sugar.Infof("Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugar.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	sugar.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		sugar.Fatalf("Server forced to shutdown: %v", err)
	}

	sugar.Info("Server exited gracefully")
}
