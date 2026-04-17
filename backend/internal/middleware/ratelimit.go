package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/digitalpapyrus/backend/internal/config"
	"github.com/digitalpapyrus/backend/pkg/response"
)

// visitor tracks rate limiting data per client.
type visitor struct {
	tokens   float64
	lastSeen time.Time
}

// RateLimiter implements a token bucket rate limiter.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     float64 // tokens per second
	burst    int     // maximum tokens
}

// NewRateLimiter creates a new token bucket rate limiter.
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     float64(requestsPerMinute) / 60.0,
		burst:    requestsPerMinute,
	}

	// Cleanup stale entries every 3 minutes
	go func() {
		for {
			time.Sleep(3 * time.Minute)
			rl.mu.Lock()
			for ip, v := range rl.visitors {
				if time.Since(v.lastSeen) > 5*time.Minute {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()

	return rl
}

// Allow checks if a request from the given key is allowed.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[key]
	if !exists {
		rl.visitors[key] = &visitor{
			tokens:   float64(rl.burst) - 1,
			lastSeen: time.Now(),
		}
		return true
	}

	// Refill tokens
	elapsed := time.Since(v.lastSeen).Seconds()
	v.tokens += elapsed * rl.rate
	if v.tokens > float64(rl.burst) {
		v.tokens = float64(rl.burst)
	}
	v.lastSeen = time.Now()

	if v.tokens < 1 {
		return false
	}

	v.tokens--
	return true
}

// RateLimitMiddleware applies general rate limiting based on client IP.
func RateLimitMiddleware(cfg *config.Config) gin.HandlerFunc {
	limiter := NewRateLimiter(cfg.Rate.General)

	return func(c *gin.Context) {
		key := c.ClientIP()
		if !limiter.Allow(key) {
			response.TooManyRequests(c, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}
		c.Next()
	}
}

// AuthRateLimitMiddleware applies stricter rate limiting for authentication endpoints.
func AuthRateLimitMiddleware(cfg *config.Config) gin.HandlerFunc {
	limiter := NewRateLimiter(cfg.Rate.Auth)

	return func(c *gin.Context) {
		key := c.ClientIP()
		if !limiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Too many login attempts. Please wait before trying again.",
				"error":   gin.H{"code": "AUTH_RATE_LIMIT"},
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
