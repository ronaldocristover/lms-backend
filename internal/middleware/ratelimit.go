package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ronaldocristover/lms-backend/pkg/apierror"
	"github.com/ronaldocristover/lms-backend/pkg/response"
)

type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int
	window   time.Duration
}

type visitor struct {
	count    int
	lastSeen time.Time
}

func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}

	go func() {
		for {
			time.Sleep(time.Minute)
			rl.cleanup()
		}
	}()

	return rl
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for ip, v := range rl.visitors {
		if time.Since(v.lastSeen) > rl.window {
			delete(rl.visitors, ip)
		}
	}
}

func (rl *RateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{lastSeen: time.Now()}
		rl.visitors[ip] = v
	}

	v.lastSeen = time.Now()
	v.count++

	return v
}

func RateLimit() gin.HandlerFunc {
	limiter := NewRateLimiter(100, time.Minute)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		v := limiter.getVisitor(ip)

		if v.count > limiter.rate {
			response.Error(c, apierror.TooManyRequests("Too many requests. Please try again later."))
			c.Abort()
			return
		}

		c.Next()
	}
}
