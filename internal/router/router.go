package router

import (
	"github.com/mohammed-r9/taxi-api/db"
	"github.com/mohammed-r9/taxi-api/internal/health"
	"github.com/mohammed-r9/taxi-api/internal/network"
	"github.com/mohammed-r9/taxi-api/internal/users"
	"github.com/redis/go-redis/v9"
)

// Deps holds the shared dependencies passed to feature packages.
type Deps struct {
	Store    *db.Store
	Redis    *redis.Client
	Validate *network.RequestValidator
}

// Register initializes the HTTP engine and wires every feature's routes.
// It is the single composition root for the API.
func Register(api *network.Api, deps Deps) {
	api.Init()

	health.Register(api)
	users.Register(api, users.Deps{
		Store:     deps.Store,
		Redis:     deps.Redis,
		Validator: deps.Validate,
	})
}
