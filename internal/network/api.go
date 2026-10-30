package network

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v5"
	"github.com/mohammed-r9/go-boiler/db"
	"github.com/redis/go-redis/v9"
)

type Api struct {
	Engine    *echo.Echo
	q         *db.Store
	rdb       *redis.Client
	validator *RequestValidator
	humaAPI   huma.API
}

func NewApi(e *echo.Echo, q *db.Store, rdb *redis.Client, v *RequestValidator) *Api {
	e.Validator = v
	return &Api{
		Engine:    e,
		q:         q,
		rdb:       rdb,
		validator: v,
	}
}
