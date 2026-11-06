package network

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rs/zerolog"
)

type requestStartKey struct{}

func RequestLogger(logger *zerolog.Logger) echo.MiddlewareFunc {
	if logger == nil {
		nop := zerolog.Nop()
		logger = &nop
	}

	requestLogger := middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
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
			if values.Error != nil {
				return nil
			}

			event := logger.WithLevel(requestLogLevel(values.Status)).
				Str("request_id", values.RequestID).
				Str("method", values.Method).
				Str("path", values.URIPath).
				Str("route", values.RoutePath).
				Int("status", values.Status).
				Str("latency", values.Latency.String()).
				Int64("response_bytes", values.ResponseSize).
				Str("ip", values.RemoteIP)
			if userID := UserIDFromContext(c); userID != "" {
				event = event.Str("user_id", userID)
			}
			event.Msg("request completed")
			return nil
		},
	})

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		withLogger := func(c *echo.Context) error {
			ctx := logger.WithContext(c.Request().Context())
			ctx = context.WithValue(ctx, requestStartKey{}, time.Now())
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}

		return requestLogger(withLogger)
	}
}

func requestLogLevel(status int) zerolog.Level {
	if status >= http.StatusInternalServerError {
		return zerolog.ErrorLevel
	}
	if status >= http.StatusBadRequest {
		return zerolog.WarnLevel
	}
	return zerolog.InfoLevel
}
