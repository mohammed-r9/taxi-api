package network

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/mohammed-r9/taxi-api/db"
	"github.com/mohammed-r9/taxi-api/internal/apperrors"
	"github.com/mohammed-r9/taxi-api/internal/auth"
	"github.com/redis/go-redis/v9"
)

const (
	rolePermissionsCachePrefix = "rbac:v1:role:permissions:"
	rolePermissionsCacheTTL    = 5 * time.Minute
)

type Policy struct {
	Attempts int64
	Window   time.Duration
}

type Middleware struct {
	store *db.Store
	rdb   *redis.Client
}

func NewMiddleware(store *db.Store, rdb *redis.Client) *Middleware {
	return &Middleware{store: store, rdb: rdb}
}

func (m *Middleware) Auth(permissions ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if header == "" {
				return apperrors.MissingAuthorizationHeader
			}

			token := strings.TrimPrefix(header, "Bearer ")
			if token == "" {
				return apperrors.InvalidToken
			}

			user, err := auth.ValidateAccessToken(token)
			if err != nil {
				return err
			}

			ctx := WithUserAccess(c.Request().Context(), user)
			c.SetRequest(c.Request().WithContext(ctx))
			SetUserID(c, strconv.FormatInt(user.UserID, 10))

			if len(permissions) > 0 {
				granted, err := m.rolePermissions(c.Request().Context(), user.RoleID)
				if err != nil {
					return err
				}
				if !hasPermission(granted, permissions) {
					return apperrors.MissingRequiredPermissions
				}
			}

			return next(c)
		}
	}
}

// RateLimit limits requests using the policy and key returned by keyFunc.
func (m *Middleware) RateLimit(policy Policy, keyFunc RateLimitKeyFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx := c.Request().Context()

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

			storedCount, err := m.rdb.Incr(ctx, key).Result()
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
				if err := m.rdb.Expire(ctx, key, policy.Window).Err(); err != nil {
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

func (m *Middleware) rolePermissions(ctx context.Context, roleID int64) ([]string, error) {
	key := rolePermissionsCacheKey(roleID)

	if m.rdb != nil {
		cached, err := m.rdb.SMembers(ctx, key).Result()
		if err == nil && len(cached) > 0 {
			return cached, nil
		}
		if err != nil && !errors.Is(err, redis.Nil) {
			slog.WarnContext(ctx, "failed to read role permissions from cache",
				"role_id", roleID,
				"error", err,
			)
		}
	}

	if m.store == nil {
		return nil, errors.New("rbac: permission store is not configured")
	}

	permissions, err := m.store.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("get role permissions: %w", err)
	}

	if m.rdb != nil && len(permissions) > 0 {
		if err := cacheRolePermissions(ctx, m.rdb, key, permissions); err != nil {
			slog.WarnContext(ctx, "failed to cache role permissions",
				"role_id", roleID,
				"error", err,
			)
		}
	}

	return permissions, nil
}

func cacheRolePermissions(ctx context.Context, rdb *redis.Client, key string, permissions []string) error {
	members := make([]interface{}, len(permissions))
	for i, permission := range permissions {
		members[i] = permission
	}

	if err := rdb.SAdd(ctx, key, members...).Err(); err != nil {
		return err
	}

	return rdb.Expire(ctx, key, rolePermissionsCacheTTL).Err()
}

func rolePermissionsCacheKey(roleID int64) string {
	return rolePermissionsCachePrefix + strconv.FormatInt(roleID, 10)
}
