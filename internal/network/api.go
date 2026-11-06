package network

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
)

type Api struct {
	Engine    *echo.Echo
	Validator *RequestValidator
	OpenAPI   *openapi3.T
	Logger    *zerolog.Logger
}

func NewApi(e *echo.Echo, v *RequestValidator) *Api {
	e.Validator = v
	return &Api{Engine: e, Validator: v}
}
