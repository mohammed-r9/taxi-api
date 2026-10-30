package network

import (
	"context"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v5"
)

// Route describes the documentation and behavior shared by an API operation.
// Echo middleware can be passed directly to Handle after the handler.
type Route struct {
	ID               string
	Summary          string
	Description      string
	Tag              string
	Status           int
	Errors           []int
	Auth             bool
	Permissions      []string
	SkipValidateBody bool
}

func Handle[I, O any](
	api *Api,
	method string,
	path string,
	route Route,
	handler func(context.Context, *I) (*O, error),
	middleware ...echo.MiddlewareFunc,
) {
	op := huma.Operation{
		OperationID:      route.ID,
		Method:           method,
		Path:             path,
		Summary:          route.Summary,
		Description:      route.Description,
		DefaultStatus:    route.Status,
		Errors:           route.Errors,
		SkipValidateBody: route.SkipValidateBody,
	}
	if route.Tag != "" {
		op.Tags = []string{route.Tag}
	}
	if route.Auth {
		op.Security = []map[string][]string{{"BearerAuth": {}}}
	}
	if len(route.Permissions) > 0 {
		required := "Required permissions: `" + strings.Join(route.Permissions, "`, `") + "`"
		if op.Description == "" {
			op.Description = required
		} else {
			op.Description += "\n\n" + required
		}
	}
	for _, mw := range middleware {
		op.Middlewares = append(op.Middlewares, echoMiddleware(api, method, path, route.ID, mw))
	}

	huma.Register(api.humaAPI, op, func(ctx context.Context, input *I) (*O, error) {
		output, err := handler(ctx, input)
		if err != nil {
			logOperationError(api.Engine.Logger, ctx, method, path, route.ID, err)
		}
		return output, err
	})
}

func echoMiddleware(api *Api, method, path, operation string, middleware echo.MiddlewareFunc) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		echoCtx := humaecho.Unwrap(ctx)
		err := middleware(func(c *echo.Context) error {
			next(huma.WithContext(ctx, c.Request().Context()))
			return nil
		})(echoCtx)
		if err == nil {
			return
		}

		response, _ := echo.UnwrapResponse(echoCtx.Response())
		if response != nil && response.Committed {
			return
		}
		logOperationError(api.Engine.Logger, echoCtx.Request().Context(), method, path, operation, err)
		api.writeHumaError(ctx, err)
	}
}

func (api *Api) writeHumaError(ctx huma.Context, err error) {
	response := toHumaError(err).(*ErrorResponse)
	_ = huma.WriteErr(api.humaAPI, ctx, response.GetStatus(), response.Message, response)
}
