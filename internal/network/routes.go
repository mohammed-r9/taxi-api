package network

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
)

func (api *Api) InitRoutesV1() {
	api.Engine.HTTPErrorHandler = ErrorHandler(api.Engine.Logger)

	config := huma.DefaultConfig("Go Boiler API", "0.1.0")
	config.DocsPath = "/api/docs"
	config.OpenAPIPath = "/api/docs/openapi"
	config.SchemasPath = "/api/docs/schemas"
	config.DocsRenderer = huma.DocsRendererScalar
	config.CreateHooks = nil
	config.Transformers = nil
	config.OpenAPI.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"BearerAuth": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "JWT access token. The Bearer prefix is added automatically.",
		},
	}

	configureHumaErrors()
	api.humaAPI = humaecho.New(api.Engine, config)
	api.registerOperations()
}

func (api *Api) registerOperations() {
	Handle(api, http.MethodGet, "/api/v1/health", Route{
		ID: "health", Summary: "Check API availability", Tag: "System",
	}, func(context.Context, *struct{}) (*healthOutput, error) {
		return &healthOutput{Body: "status is available"}, nil
	})
}

type healthOutput struct {
	Body string `contentType:"text/plain"`
}
