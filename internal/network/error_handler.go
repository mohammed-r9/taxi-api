package network

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/mohammed-r9/taxi-api/internal/apperrors"
	"github.com/rs/zerolog"
)

type errorInfo struct {
	code    string
	message string
	detail  []string
}

const userIDContextKey = "user_id"

func UserIDFromContext(c *echo.Context) string {
	if id, ok := c.Get(userIDContextKey).(string); ok {
		return id
	}
	return ""
}

// SetUserID stores the request's user identity for logging and rate limiting.
// Callers may set a validated challenge identity before authentication; it does
// not grant access to protected routes.
func SetUserID(c *echo.Context, id string) {
	c.Set(userIDContextKey, id)
}

func ErrorHandler(logger *zerolog.Logger) echo.HTTPErrorHandler {
	if logger == nil {
		nop := zerolog.Nop()
		logger = &nop
	}

	return func(c *echo.Context, err error) {
		resp, _ := echo.UnwrapResponse(c.Response())
		if resp != nil && resp.Committed {
			return
		}

		status, info := mapError(err)

		event := logger.WithLevel(levelFor(status)).
			Str("route", c.Path()).
			Str("method", c.Request().Method).
			Int("status", status).
			Str("error", err.Error()).
			Str("ip", c.RealIP()).
			Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID))

		if start, ok := c.Request().Context().Value(requestStartKey{}).(time.Time); ok {
			event = event.Str("latency", time.Since(start).String())
		}

		if userID := UserIDFromContext(c); userID != "" {
			event = event.Str("user_id", userID)
		}

		event.Msg("request failed")

		if c.Request().Method == http.MethodHead {
			_ = c.NoContent(status)
			return
		}

		if len(info.detail) > 0 {
			_ = WriteValidationError(
				c,
				status,
				info.code,
				info.message,
				info.detail,
			)
			return
		}

		_ = WriteError(c, status, info.code, info.message)
	}
}

func levelFor(status int) zerolog.Level {
	if status >= http.StatusInternalServerError {
		return zerolog.ErrorLevel
	}
	return zerolog.WarnLevel
}

func mapError(err error) (int, errorInfo) {
	if validationErr, ok := errors.AsType[*RequestValidationError](err); ok {
		return http.StatusBadRequest, errorInfo{
			code:    "VALIDATION_ERROR",
			message: "request validation failed",
			detail:  validationErr.Messages,
		}
	}

	if appErr, ok := errors.AsType[apperrors.AppError](err); ok {
		return appErr.Status(), errorInfo{
			code:    appErr.Code(),
			message: appErr.Error(),
		}
	}

	if httpErr, ok := errors.AsType[*echo.HTTPError](err); ok {
		status := httpErr.Code
		message := httpErr.Message

		if message == "" {
			message = http.StatusText(status)
		}

		return status, errorInfo{
			message: message,
		}
	}

	return http.StatusInternalServerError, errorInfo{
		code:    "internal_error",
		message: http.StatusText(http.StatusInternalServerError),
	}
}
