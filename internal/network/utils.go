package network

import (
	"context"
	"slices"

	"github.com/labstack/echo/v5"
)

type Map map[string]any

type RateLimitKeyFunc func(c *echo.Context) (string, error)

func RateLimitByUserID(c *echo.Context) (string, error) {
	return "rate_limit:" + c.Request().Method + ":" + c.Path() + ":id:" + UserIDFromContext(c), nil
}

func RateLimitByIP(c *echo.Context) (string, error) {
	return "rate_limit:" + c.Request().Method + ":" + c.Path() + ":ip:" + c.RealIP(), nil
}

func hasPermission(perms, allowed []string) bool {
	for _, p := range perms {
		if slices.Contains(allowed, p) {
			return true
		}
	}
	return false
}

type userAccessContextKey struct{}

func WithUserAccess(ctx context.Context, user any) context.Context {
	return context.WithValue(ctx, userAccessContextKey{}, user)
}

func UserAccessFromContext(ctx context.Context) (any, bool) {
	user, ok := ctx.Value(userAccessContextKey{}).(any)
	return user, ok
}
