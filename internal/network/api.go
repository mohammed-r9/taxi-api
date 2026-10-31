package network

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v5"
)

type Api struct {
	Engine  *echo.Echo
	humaAPI huma.API
}

func NewApi(e *echo.Echo, v *RequestValidator) *Api {
	e.Validator = v
	return &Api{Engine: e}
}
