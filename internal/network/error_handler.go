package network

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/mohammed-r9/go-boiler/internal/apperrors"
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

func ErrorHandler(logger *slog.Logger) echo.HTTPErrorHandler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(c *echo.Context, err error) {
		resp, _ := echo.UnwrapResponse(c.Response())
		if resp != nil && resp.Committed {
			return
		}

		status, info := mapError(err)

		args := []any{
			slog.String("route", c.Path()),
			slog.String("method", c.Request().Method),
			slog.Int("status", status),
			slog.String("error", err.Error()),
			slog.String("ip", c.RealIP()),
		}

		if userID := UserIDFromContext(c); userID != "" {
			args = append(args, slog.String("user_id", userID))
		}

		logger.Log(
			c.Request().Context(),
			levelFor(status),
			"request failed",
			args...,
		)

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

func levelFor(status int) slog.Level {
	if status >= 500 {
		return slog.LevelError
	}
	return slog.LevelWarn
}
