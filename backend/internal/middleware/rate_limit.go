package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/util"
	"github.com/redis/go-redis/v9"
)

// RateLimitMiddleware Redis 限流中间件（漏桶计数）。
type RateLimitMiddleware struct {
	client *redis.Client
	logger *slog.Logger
	limit  int
	window time.Duration
}

// NewRateLimit 构造限流中间件。
func NewRateLimit(client *redis.Client, logger *slog.Logger, limit int, windowSeconds int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		client: client,
		logger: logger,
		limit:  limit,
		window: time.Duration(windowSeconds) * time.Second,
	}
}

// Limit 返回限流中间件；Redis 不可用时放行（fail-open）。
func (m *RateLimitMiddleware) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "paperflow:rl:" + c.ClientIP()
		ctx := c.Request.Context()
		count, err := m.client.Incr(ctx, key).Result()
		if err != nil {
			m.logger.Warn("rate limit redis error, fail open", "error", err)
			c.Next()
			return
		}
		if count == 1 {
			m.client.Expire(ctx, key, m.window)
		}
		if count > int64(m.limit) {
			m.logger.Warn(fmt.Sprintf(constants.LogRateLimited, c.ClientIP()))
			util.Fail(c, http.StatusTooManyRequests, constants.CodeTooManyRequests, constants.MsgRateLimited)
			return
		}
		c.Next()
	}
}
