package network

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// RequestLogger logs request metadata, but never bodies or credentials.
func RequestLogger(logger *slog.Logger) echo.MiddlewareFunc {
	if logger == nil {
		logger = slog.Default()
	}

	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		HandleError:     true,
		LogLatency:      true,
		LogRemoteIP:     true,
		LogMethod:       true,
		LogURIPath:      true,
		LogRoutePath:    true,
		LogRequestID:    true,
		LogStatus:       true,
		LogResponseSize: true,
		LogValuesFunc: func(c *echo.Context, values middleware.RequestLoggerValues) error {
			attrs := []any{
				slog.String("request_id", values.RequestID),
				slog.String("method", values.Method),
				slog.String("path", values.URIPath),
				slog.String("route", values.RoutePath),
				slog.Int("status", values.Status),
				slog.Duration("latency", values.Latency),
				slog.Int64("response_bytes", values.ResponseSize),
				slog.String("ip", values.RemoteIP),
			}
			if userID := UserIDFromContext(c); userID != "" {
				attrs = append(attrs, slog.String("user_id", userID))
			}
			if values.Error != nil {
				attrs = append(attrs, slog.String("error", values.Error.Error()))
			}
			logger.Log(c.Request().Context(), requestLogLevel(values.Status), "request completed", attrs...)
			return nil
		},
	})
}

func requestLogLevel(status int) slog.Level {
	if status >= http.StatusInternalServerError {
		return slog.LevelError
	}
	if status >= http.StatusBadRequest {
		return slog.LevelWarn
	}
	return slog.LevelInfo
}

func logOperationError(logger *slog.Logger, ctx context.Context, method, path, operation string, err error) {
	if logger == nil {
		logger = slog.Default()
	}
	status, info := mapError(err)
	cause := err
	if response, ok := err.(*ErrorResponse); ok && response.cause != nil {
		cause = response.cause
	}
	attrs := []any{
		slog.String("operation", operation),
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", status),
		slog.String("code", info.code),
		slog.String("error", err.Error()),
		slog.String("cause", cause.Error()),
	}
	if userID, ok := ctx.Value(userIDContextKey).(string); ok && userID != "" {
		attrs = append(attrs, slog.String("user_id", userID))
	}
	logger.Log(ctx, requestLogLevel(status), "api operation failed", attrs...)
}
