package network

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

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
