package server

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, int64, error)
	GetTokens(ctx context.Context, key string) (float64, error)
}

type KeyExtractor func(c *gin.Context) string

func IPKeyExtractor() KeyExtractor {
	return func(c *gin.Context) string {
		return c.ClientIP()
	}
}

func APIKeyExtractor(headerName string) KeyExtractor {
	return func(c *gin.Context) string {
		if apiKey := c.GetHeader(headerName); apiKey != "" {
			return apiKey
		}
		return "ip:" + c.ClientIP()
	}
}

type RateLimitConfig struct {
	KeyExtractor   KeyExtractor
	ErrorMessage   string
	IncludeHeaders bool
	Limit          float64
	RefillRate     float64
	FailOpen       bool
}

func DefaultConfig() RateLimitConfig {
	return RateLimitConfig{
		KeyExtractor:   IPKeyExtractor(),
		ErrorMessage:   "Rate limit exceeded. Please try again later",
		IncludeHeaders: true,
		Limit:          100,
		RefillRate:     10,
		FailOpen:       true,
	}
}

func (s *Server) RateLimiterMiddleware(limiter RateLimiter, config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		key := config.KeyExtractor(c)
		allowed, remaining, err := limiter.Allow(ctx, key)
		if err != nil {
			// Log error but allow request to proceed or stop based on failOpen config
			c.Header("X-RateLimit-Error", "service_unavailable")
			if config.FailOpen {
				c.Next()
			}
			return
		}
		if !allowed {
			if config.IncludeHeaders {
				c.Header("X-RateLimit-Limit", fmt.Sprintf("%.1f", config.Limit))
				c.Header("X-RateLimit-Remaining", "0")
				c.Header("Retry-After", strconv.Itoa(int(math.Ceil(1/config.RefillRate))))
			}

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": config.ErrorMessage,
			})
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%.1f", config.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Next()
	}
}
