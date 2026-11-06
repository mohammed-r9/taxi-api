package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/mohammed-r9/taxi-api/db"
	"github.com/mohammed-r9/taxi-api/internal/network"
	"github.com/mohammed-r9/taxi-api/internal/router"
	"github.com/rs/zerolog"
)

const shutdownTimeout = 5 * time.Second

type application struct {
	addr   string
	api    *network.Api
	db     *pgxpool.Pool
	logger *zerolog.Logger
}

func (a *application) run() error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	server := echo.StartConfig{
		Address:         a.addr,
		GracefulTimeout: shutdownTimeout,
		OnShutdownError: func(err error) {
			a.logger.Error().Err(err).Msg("graceful shutdown failed")
		},
	}

	a.logger.Info().Str("addr", a.addr).Msg("server started")

	err := server.Start(ctx, a.api.Engine)

	if ctx.Err() != nil {
		a.logger.Info().Msg("shutdown signal received")
	}

	a.shutdown()

	return err
}

func (a *application) shutdown() {
	if a.db != nil {
		a.logger.Info().Msg("closing postgres pool...")
		a.db.Close()
	}

	a.logger.Info().Msg("server stopped")
}

func (a *application) mount() {
	logger := newLogger()
	a.logger = logger

	e := echo.New()
	e.Logger = slog.New(zerolog.NewSlogHandler(*logger))
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return uuid.NewV4().String()
		},
	}))
	e.Use(network.RequestLogger(logger))
	e.Use(middleware.Recover())
	e.Use(middleware.CORS("*"))

	postgres := db.InitDB()
	redisdb := db.NewRedisClient(db.RedisConfig{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
	})
	store := db.NewStore(postgres)
	validator := network.NewRequestValidator()

	a.db = postgres
	a.api = network.NewApi(e, validator)
	a.api.Logger = logger
	router.Register(a.api, router.Deps{
		Store:    store,
		Redis:    redisdb,
		Validate: validator,
	})
}

func newLogger() *zerolog.Logger {
	var out io.Writer = os.Stdout
	if os.Getenv("LOCAL") == "true" {
		out = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	}

	logger := zerolog.New(out).With().Timestamp().Logger()
	return &logger
}
