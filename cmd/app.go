package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/mohammed-r9/go-boiler/db"
	"github.com/mohammed-r9/go-boiler/internal/network"
)

const shutdownTimeout = 5 * time.Second

type application struct {
	addr string
	api  *network.Api
	db   *pgxpool.Pool
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
			log.Printf("graceful shutdown failed: %v", err)
		},
	}

	log.Printf("server started at %s", a.addr)

	err := server.Start(ctx, a.api.Engine)

	if ctx.Err() != nil {
		log.Println("shutdown signal received")
	}

	a.shutdown()

	return err
}

func (a *application) shutdown() {
	if a.db != nil {
		log.Println("closing postgres pool...")
		a.db.Close()
	}

	log.Println("server stopped")
}

func (a *application) mount() {
	e := echo.New()
	if os.Getenv("LOCAL") == "true" {
		e.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	e.Use(middleware.RequestID())
	e.Use(network.RequestLogger(e.Logger))
	e.Use(middleware.Recover())
	e.Use(middleware.CORS("*"))

	e.HTTPErrorHandler = network.ErrorHandler(e.Logger)

	postgres := db.InitDB()
	redisdb := db.NewRedisClient(db.RedisConfig{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
	})
	store := db.NewStore(postgres)
	validator := network.NewRequestValidator()

	a.db = postgres
	a.api = network.NewApi(e, store, redisdb, validator)
	a.api.InitRoutesV1()
}
