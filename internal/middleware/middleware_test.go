package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupMiddlewareTest() (*gin.Engine, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	w := httptest.NewRecorder()
	return r, w
}

func TestAuth_MissingHeader(t *testing.T) {
	r, w := setupMiddlewareTest()
	r.Use(Auth("secret"))
	r.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Missing authorization header")
}

func TestAuth_InvalidHeaderFormat(t *testing.T) {
	r, w := setupMiddlewareTest()
	r.Use(Auth("secret"))
	r.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid authorization header format")
}

func TestAuth_InvalidToken(t *testing.T) {
	r, w := setupMiddlewareTest()
	r.Use(Auth("secret"))
	r.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

func TestAuth_ValidToken(t *testing.T) {
	r, w := setupMiddlewareTest()
	r.Use(Auth("secret"))
	r.GET("/protected", func(c *gin.Context) {
		userID, _ := c.Get("userID")
		c.String(http.StatusOK, userID.(uuid.UUID).String())
	})

	userID := uuid.New()
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   "test@example.com",
		"role":    "user",
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("secret"))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), userID.String())
}

func TestAuth_ExpiredToken(t *testing.T) {
	r, w := setupMiddlewareTest()
	r.Use(Auth("secret"))
	r.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	claims := jwt.MapClaims{
		"user_id": uuid.New(),
		"email":   "test@example.com",
		"role":    "user",
		"exp":     time.Now().Add(-time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("secret"))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_WrongSecret(t *testing.T) {
	r, w := setupMiddlewareTest()
	r.Use(Auth("wrong-secret"))
	r.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	claims := jwt.MapClaims{
		"user_id": uuid.New(),
		"email":   "test@example.com",
		"role":    "user",
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("secret"))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCORS(t *testing.T) {
	r, w := setupMiddlewareTest()
	r.Use(CORS(CORSConfig{}))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Origin, Content-Type, Accept, Authorization, X-Request-ID", w.Header().Get("Access-Control-Allow-Headers"))
}

func TestCORS_Preflight(t *testing.T) {
	r, w := setupMiddlewareTest()
	r.Use(CORS(CORSConfig{}))
	r.OPTIONS("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestRequestID(t *testing.T) {
	r, w := setupMiddlewareTest()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestRequestID_Custom(t *testing.T) {
	r, w := setupMiddlewareTest()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "custom-id-123")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "custom-id-123", w.Header().Get("X-Request-ID"))
}

func TestRecovery(t *testing.T) {
	r, w := setupMiddlewareTest()
	sugar := zap.NewNop().Sugar()
	r.Use(Recovery(sugar))
	r.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRecovery_NoPanic(t *testing.T) {
	r, w := setupMiddlewareTest()
	sugar := zap.NewNop().Sugar()
	r.Use(Recovery(sugar))
	r.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogger(t *testing.T) {
	r, w := setupMiddlewareTest()
	sugar := zap.NewNop().Sugar()
	r.Use(Logger(sugar))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test?foo=bar", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimit_Allow(t *testing.T) {
	r, w := setupMiddlewareTest()
	r.Use(RateLimit())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		r.ServeHTTP(w, req)
	}

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimit_Exceed(t *testing.T) {
	r, w := setupMiddlewareTest()
	r.Use(RateLimit())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	for i := 0; i < 101; i++ {
		w = httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		r.ServeHTTP(w, req)
	}

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Too many requests")
}

func TestRateLimit_DifferentIPs(t *testing.T) {
	r, w := setupMiddlewareTest()
	r.Use(RateLimit())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	for i := 0; i < 150; i++ {
		w = httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = fmt.Sprintf("10.0.0.%d:1234", i%10)
		r.ServeHTTP(w, req)
	}

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, time.Second)
	assert.NotNil(t, rl)
	assert.Equal(t, 10, rl.rate)
	assert.Equal(t, time.Second, rl.window)
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := NewRateLimiter(5, 100*time.Millisecond)

	for i := 0; i < 3; i++ {
		_ = rl.getVisitor(fmt.Sprintf("10.0.0.%d", i))
	}

	assert.Equal(t, 3, len(rl.visitors))

	time.Sleep(200 * time.Millisecond)
	rl.cleanup()

	assert.Equal(t, 0, len(rl.visitors))
}

func TestRateLimiter_VisitorCount(t *testing.T) {
	rl := NewRateLimiter(100, time.Minute)

	v1 := rl.getVisitor("1.2.3.4")
	assert.Equal(t, 1, v1.count)

	v1 = rl.getVisitor("1.2.3.4")
	assert.Equal(t, 2, v1.count)

	v2 := rl.getVisitor("5.6.7.8")
	assert.Equal(t, 1, v2.count)
}
