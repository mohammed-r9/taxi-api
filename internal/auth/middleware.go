package auth

import (
	"context"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/mohammed-r9/taxi-api/internal/apperrors"
	"github.com/mohammed-r9/taxi-api/internal/network"
)

type userContextKey struct{}

// WithUser stores the authenticated user on a request context.
func WithUser(ctx context.Context, user UserAccessData) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

// UserFromContext returns the authenticated user stored on ctx.
func UserFromContext(ctx context.Context) (UserAccessData, bool) {
	user, ok := ctx.Value(userContextKey{}).(UserAccessData)
	return user, ok
}

// RequireAuth validates the Bearer access token and stores the user on the
// request context.
func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		header := c.Request().Header.Get("Authorization")
		if header == "" {
			return apperrors.MissingAuthorizationHeader
		}

		scheme, token, ok := strings.Cut(header, " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
			return apperrors.InvalidToken
		}

		user, err := ValidateAccessToken(token)
		if err != nil {
			return err
		}

		ctx := WithUser(c.Request().Context(), user)
		c.SetRequest(c.Request().WithContext(ctx))
		network.SetUserID(c, strconv.FormatInt(user.UserID, 10))

		return next(c)
	}
}
