package network

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	rdb *redis.Client
}

type Policy struct {
	Attempts int64
	Window   time.Duration
}

func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{rdb: rdb}
}

func (l *RateLimiter) Middleware(policy Policy, keyFunc RateLimitKeyFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx := c.Request().Context()

			// key := fmt.Sprintf(
			// 	"rate_limit:%s:%s:%s",
			// 	c.Request().Method,
			// 	c.Path(),
			// 	c.RealIP(),
			// )
			key, err := keyFunc(c)
			if err != nil {
				slog.ErrorContext(ctx, "failed to build rate-limit key",
					"error", err,
					"key", key,
				)

				return WriteError(
					c,
					http.StatusInternalServerError,
					"RATE_LIMIT_ERROR",
					"failed to check rate limit",
					err,
				)
			}

			storedCount, err := l.rdb.Incr(ctx, key).Result()
			if err != nil {
				slog.ErrorContext(ctx, "failed to increment rate-limit counter",
					"error", err,
					"key", key,
				)

				return WriteError(
					c,
					http.StatusInternalServerError,
					"RATE_LIMIT_ERROR",
					"failed to check rate limit",
					err,
				)
			}

			if storedCount == 1 {
				if err := l.rdb.Expire(ctx, key, policy.Window).Err(); err != nil {
					slog.ErrorContext(ctx, "failed to set rate-limit expiration",
						"error", err,
						"key", key,
					)

					return WriteError(
						c,
						http.StatusInternalServerError,
						"RATE_LIMIT_ERROR",
						"failed to check rate limit",
						err,
					)
				}
			}

			if storedCount > policy.Attempts {
				return WriteError(
					c,
					http.StatusTooManyRequests,
					"RATE_LIMIT_EXCEEDED",
					"too many requests",
				)
			}

			return next(c)
		}
	}
}
