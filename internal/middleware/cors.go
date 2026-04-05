package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

func CORS(cfg CORSConfig) gin.HandlerFunc {
	defaultMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	defaultHeaders := []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"}

	methods := cfg.AllowedMethods
	if len(methods) == 0 {
		methods = defaultMethods
	}

	headers := cfg.AllowedHeaders
	if len(headers) == 0 {
		headers = defaultHeaders
	}

	maxAge := cfg.MaxAge
	if maxAge == 0 {
		maxAge = 86400 // 24 hours
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if origin != "" && isOriginAllowed(origin, cfg.AllowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if len(cfg.AllowedOrigins) == 0 {
			// Dev mode: allow all if no origins configured
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Methods", strings.Join(methods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(headers, ", "))
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")

		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.Header("Access-Control-Max-Age", http.StatusText(maxAge))

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func isOriginAllowed(origin string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, o := range allowed {
		if o == "*" {
			return true
		}
		if o == origin {
			return true
		}
	}
	return false
}
