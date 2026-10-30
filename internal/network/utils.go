package network

import (
	"slices"

	"github.com/labstack/echo/v5"
)

type Map map[string]any

type RateLimitKeyFunc func(c *echo.Context) (string, error)

func RateLimitByIP(c *echo.Context) (string, error) {
	return "ip:" + c.RealIP(), nil
}

func RateLimitByUserID(c *echo.Context) (string, error) {
	return "id:" + UserIDFromContext(c), nil
}

func hasPermission(perms, allowed []string) bool {
	for _, p := range perms {
		if slices.Contains(allowed, p) {
			return true
		}
	}
	return false
}
