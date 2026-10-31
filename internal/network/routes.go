package network

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
)

func (api *Api) Init() {
	api.Engine.HTTPErrorHandler = ErrorHandler(api.Engine.Logger)

	config := huma.DefaultConfig("Taxi API", "0.1.0")
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
}
