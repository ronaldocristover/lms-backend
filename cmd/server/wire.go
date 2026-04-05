package main

import (
	"fmt"

	"github.com/ronaldocristover/lms-backend/internal/config"
	"github.com/ronaldocristover/lms-backend/internal/handler"
	"github.com/ronaldocristover/lms-backend/internal/model"
	"github.com/ronaldocristover/lms-backend/internal/repository"
	"github.com/ronaldocristover/lms-backend/internal/scheduler"
	"github.com/ronaldocristover/lms-backend/internal/service"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type handlers struct {
	Auth      *handler.AuthHandler
	User      *handler.UserHandler
	Role      *handler.RoleHandler
	Org       *handler.OrganizationHandler
	Upload    *handler.UploadHandler
	Language  *handler.LanguageHandler
	Media     *handler.MediaHandler
	Subtitle  *handler.SubtitleHandler
	Category  *handler.CategoryHandler
	Series    *handler.SeriesHandler
	Session   *handler.SessionHandler
	Content   *handler.ContentHandler
	Health    *handler.HealthHandler
	Scheduler *scheduler.Scheduler
}

func initDB(cfg *config.Config, sugar *zap.SugaredLogger) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.DBName,
	)

	gormConfig := &gorm.Config{}
	if cfg.Server.Env != "production" {
		gormConfig.Logger = gormlogger.Default.LogMode(gormlogger.Info)
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

	if err := db.AutoMigrate(&model.Role{}, &model.User{}, &model.Organization{}, &model.OrganizationUser{}, &model.Language{}, &model.Media{}, &model.Subtitle{}, &model.Category{}, &model.Series{}, &model.Session{}, &model.Content{}); err != nil {
		sugar.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

func setupServices(cfg *config.Config, db *gorm.DB, sugar *zap.SugaredLogger) *handlers {
	roleRepo := repository.NewRoleRepository(db)
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo, roleRepo, cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.RefreshExpiry)
	roleSvc := service.NewRoleService(roleRepo)

	orgRepo := repository.NewOrganizationRepository(db)
	orgUserRepo := repository.NewOrganizationUserRepository(db)
	orgSvc := service.NewOrganizationService(orgRepo, orgUserRepo, userRepo)

	langRepo := repository.NewLanguageRepository(db)
	langSvc := service.NewLanguageService(langRepo)

	mediaRepo := repository.NewMediaRepository(db)
	mediaSvc := service.NewMediaService(mediaRepo, langRepo)

	subtitleRepo := repository.NewSubtitleRepository(db)
	subtitleSvc := service.NewSubtitleService(subtitleRepo, mediaRepo, langRepo)

	catRepo := repository.NewCategoryRepository(db)
	catSvc := service.NewCategoryService(catRepo)

	seriesRepo := repository.NewSeriesRepository(db)
	seriesSvc := service.NewSeriesService(seriesRepo, catRepo)

	sessionRepo := repository.NewSessionRepository(db)
	sessionSvc := service.NewSessionService(sessionRepo, seriesRepo)

	contentRepo := repository.NewContentRepository(db)
	contentSvc := service.NewContentService(contentRepo, sessionRepo)

	sched := scheduler.NewScheduler(cfg.Jobs.Workers, cfg.Jobs.QueueSize)
	sched.Start()

	return &handlers{
		Auth:      handler.NewAuthHandler(userSvc),
		User:      handler.NewUserHandler(userSvc),
		Role:      handler.NewRoleHandler(roleSvc),
		Org:       handler.NewOrganizationHandler(orgSvc),
		Upload:    handler.NewUploadHandler(cfg.Upload.Dir, cfg.Upload.MaxSize),
		Language:  handler.NewLanguageHandler(langSvc),
		Media:     handler.NewMediaHandler(mediaSvc),
		Subtitle:  handler.NewSubtitleHandler(subtitleSvc),
		Category:  handler.NewCategoryHandler(catSvc),
		Series:    handler.NewSeriesHandler(seriesSvc),
		Session:   handler.NewSessionHandler(sessionSvc),
		Content:   handler.NewContentHandler(contentSvc),
		Health:    handler.NewHealthHandler(db, "1.0.0"),
		Scheduler: sched,
	}
}
