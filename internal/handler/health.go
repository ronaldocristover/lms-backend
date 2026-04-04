package handler

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db        *gorm.DB
	startTime time.Time
	version   string
}

type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Uptime    string            `json:"uptime"`
	Version   string            `json:"version"`
	Checks    map[string]Check  `json:"checks"`
	System    SystemInfo        `json:"system"`
}

type Check struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
}

type DetailedHealthStatus struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Uptime    string            `json:"uptime"`
	Version   string            `json:"version"`
	Checks    map[string]Check  `json:"checks"`
	System    SystemInfo        `json:"system"`
	Database  DatabaseInfo      `json:"database,omitempty"`
}

type DatabaseInfo struct {
	Status      string `json:"status"`
	Host        string `json:"host"`
	Database    string `json:"database"`
	MaxOpen     int    `json:"max_open"`
	MaxIdle     int    `json:"max_idle"`
	OpenConns   int    `json:"open_conns"`
	IdleConns   int    `json:"idle_conns"`
	Latency     string `json:"latency"`
}

func NewHealthHandler(db *gorm.DB, version string) *HealthHandler {
	return &HealthHandler{
		db:        db,
		startTime: time.Now(),
		version:   version,
	}
}

// HealthCheck godoc
// @Summary      Basic health check
// @Description  Returns OK if service is running
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// HealthStatusCheck godoc
// @Summary      Comprehensive health status
// @Description  Returns health status with database check and system info
// @Tags         health
// @Produce      json
// @Success      200  {object}  HealthStatus
// @Failure      503  {object}  HealthStatus
// @Router       /health/status [get]
func (h *HealthHandler) HealthStatusCheck(c *gin.Context) {
	checks := make(map[string]Check)
	allHealthy := true

	// Database check
	dbCheck := h.checkDatabase()
	checks["database"] = dbCheck
	if dbCheck.Status != "healthy" {
		allHealthy = false
	}

	status := "healthy"
	if !allHealthy {
		status = "unhealthy"
	}

	health := HealthStatus{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Uptime:    time.Since(h.startTime).String(),
		Version:   h.version,
		Checks:    checks,
		System: SystemInfo{
			GoVersion:    runtime.Version(),
			NumGoroutine: runtime.NumGoroutine(),
			NumCPU:       runtime.NumCPU(),
			OS:           runtime.GOOS,
			Arch:         runtime.GOARCH,
		},
	}

	statusCode := http.StatusOK
	if !allHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}

// HealthDetailedCheck godoc
// @Summary      Detailed health check
// @Description  Returns detailed health with connection pool stats
// @Tags         health
// @Produce      json
// @Success      200  {object}  DetailedHealthStatus
// @Failure      503  {object}  DetailedHealthStatus
// @Router       /health/detailed [get]
func (h *HealthHandler) HealthDetailedCheck(c *gin.Context) {
	checks := make(map[string]Check)
	allHealthy := true

	// Database check with details
	dbCheck := h.checkDatabase()
	checks["database"] = dbCheck
	if dbCheck.Status != "healthy" {
		allHealthy = false
	}

	status := "healthy"
	if !allHealthy {
		status = "unhealthy"
	}

	// Get database stats
	var dbInfo DatabaseInfo
	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err == nil {
			stats := sqlDB.Stats()
			dbInfo = DatabaseInfo{
				Status:    dbCheck.Status,
				Host:      "postgres",
				Database:  "lms",
				MaxOpen:   stats.MaxOpenConnections,
				MaxIdle:   int(stats.MaxIdleClosed),
				OpenConns: stats.OpenConnections,
				IdleConns: stats.Idle,
				Latency:   dbCheck.Latency,
			}
		}
	}

	health := DetailedHealthStatus{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Uptime:    time.Since(h.startTime).String(),
		Version:   h.version,
		Checks:    checks,
		System: SystemInfo{
			GoVersion:    runtime.Version(),
			NumGoroutine: runtime.NumGoroutine(),
			NumCPU:       runtime.NumCPU(),
			OS:           runtime.GOOS,
			Arch:         runtime.GOARCH,
		},
		Database: dbInfo,
	}

	statusCode := http.StatusOK
	if !allHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}

// HealthLiveCheck godoc
// @Summary      Liveness probe
// @Description  Kubernetes liveness probe endpoint
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health/live [get]
func (h *HealthHandler) HealthLiveCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

// HealthReadyCheck godoc
// @Summary      Readiness probe
// @Description  Kubernetes readiness probe — checks database connectivity
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      503  {object}  map[string]string
// @Router       /health/ready [get]
func (h *HealthHandler) HealthReadyCheck(c *gin.Context) {
	// Check if database is accessible
	dbCheck := h.checkDatabase()

	if dbCheck.Status != "healthy" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  dbCheck.Error,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

func (h *HealthHandler) checkDatabase() Check {
	start := time.Now()

	if h.db == nil {
		return Check{
			Status: "unhealthy",
			Error:  "database connection is nil",
		}
	}

	sqlDB, err := h.db.DB()
	if err != nil {
		return Check{
			Status: "unhealthy",
			Error:  err.Error(),
		}
	}

	ctx, cancel := DefaultContextTimeout()
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return Check{
			Status: "unhealthy",
			Error:  err.Error(),
		}
	}

	latency := time.Since(start)

	return Check{
		Status:  "healthy",
		Latency: latency.String(),
	}
}

func DefaultContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}
