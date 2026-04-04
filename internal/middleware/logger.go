package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logger(sugar *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		duration := time.Since(start)

		sugar.Infow(
			"HTTP request",
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"status", c.Writer.Status(),
			"duration", duration.String(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"request_id", c.GetHeader("X-Request-ID"),
		)
	}
}
