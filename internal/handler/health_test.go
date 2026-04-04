package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupHealthHandler(t *testing.T) (*HealthHandler, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Don't migrate User model since SQLite doesn't support gen_random_uuid()
	// We just need a working DB connection for health checks
	handler := NewHealthHandler(db, "1.0.0-test")
	return handler, db
}

func setupHealthRouter(h *HealthHandler) *gin.Engine {
	r := gin.New()
	r.GET("/health", h.HealthCheck)
	r.GET("/health/status", h.HealthStatusCheck)
	r.GET("/health/detailed", h.HealthDetailedCheck)
	r.GET("/health/live", h.HealthLiveCheck)
	r.GET("/health/ready", h.HealthReadyCheck)
	return r
}

// ─── Basic Health Check ───

func TestHealthCheck_Basic(t *testing.T) {
	handler, _ := setupHealthHandler(t)
	router := setupHealthRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "ok", resp["status"])
}

// ─── Health Status Check ───

func TestHealthStatus_Healthy(t *testing.T) {
	handler, _ := setupHealthHandler(t)
	router := setupHealthRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/health/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp HealthStatus
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, "healthy", resp.Status)
	assert.Equal(t, "1.0.0-test", resp.Version)
	assert.NotEmpty(t, resp.Timestamp)
	assert.NotEmpty(t, resp.Uptime)

	// Check database
	assert.Equal(t, "healthy", resp.Checks["database"].Status)
	assert.NotEmpty(t, resp.Checks["database"].Latency)

	// Check system info
	assert.Equal(t, runtime.Version(), resp.System.GoVersion)
	assert.Greater(t, resp.System.NumCPU, 0)
	assert.Greater(t, resp.System.NumGoroutine, 0)
	assert.NotEmpty(t, resp.System.OS)
	assert.NotEmpty(t, resp.System.Arch)
}

func TestHealthStatus_UptimeTracking(t *testing.T) {
	handler, _ := setupHealthHandler(t)

	// Set start time to 1 hour ago
	handler.startTime = time.Now().Add(-1 * time.Hour)

	router := setupHealthRouter(handler)
	req := httptest.NewRequest(http.MethodGet, "/health/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp HealthStatus
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Uptime should be approximately 1 hour
	assert.Contains(t, resp.Uptime, "1h")
}

// ─── Detailed Health Check ───

func TestHealthDetailed_Healthy(t *testing.T) {
	handler, _ := setupHealthHandler(t)
	router := setupHealthRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/health/detailed", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp DetailedHealthStatus
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, "healthy", resp.Status)
	assert.Equal(t, "1.0.0-test", resp.Version)
	assert.NotEmpty(t, resp.Timestamp)
	assert.NotEmpty(t, resp.Uptime)

	// Check database info
	assert.Equal(t, "healthy", resp.Database.Status)
	assert.NotEmpty(t, resp.Database.Latency)
	assert.GreaterOrEqual(t, resp.Database.OpenConns, 0)
	assert.GreaterOrEqual(t, resp.Database.IdleConns, 0)
}

// ─── Liveness Probe ───

func TestHealthLiveness(t *testing.T) {
	handler, _ := setupHealthHandler(t)
	router := setupHealthRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "alive", resp["status"])
}

// ─── Readiness Probe ───

func TestHealthReadiness_Ready(t *testing.T) {
	handler, _ := setupHealthHandler(t)
	router := setupHealthRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ready", resp["status"])
}

func TestHealthReady_NilDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHealthHandler(nil, "1.0.0-test")
	router := setupHealthRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "not ready", resp["status"])
}

// ─── Database Check ───

func TestDatabaseCheck_NilDB(t *testing.T) {
	handler := NewHealthHandler(nil, "1.0.0-test")
	check := handler.checkDatabase()

	assert.Equal(t, "unhealthy", check.Status)
	assert.Equal(t, "database connection is nil", check.Error)
}

func TestDatabaseCheck_ClosedConnection(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err)

	// Close the underlying connection
	sqlDB, _ := db.DB()
	sqlDB.Close()

	handler := NewHealthHandler(db, "1.0.0-test")
	check := handler.checkDatabase()

	assert.Equal(t, "unhealthy", check.Status)
}

// ─── Version ───

func TestHealthVersion(t *testing.T) {
	handler, _ := setupHealthHandler(t)
	router := setupHealthRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/health/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp HealthStatus
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "1.0.0-test", resp.Version)
}

// ─── Multiple Checks ───

func TestHealthStatus_MultipleRequests(t *testing.T) {
	handler, _ := setupHealthHandler(t)
	router := setupHealthRouter(handler)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health/status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}
}
