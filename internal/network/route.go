package network

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/mohammed-r9/taxi-api/internal/apperrors"
)

// Route describes the documentation and behavior shared by an API operation.
// Echo middleware can be passed directly to Handle after the handler.
type Route struct {
	ID          string
	Summary     string
	Description string
	Tag         string
	Status      int
	ContentType string
	Errors      []int
	Auth        bool
	Permissions []string
}

// Handler is the shape every typed API handler has. Input is bound from the
// request and Output is written to the response by Handle
type Handler[I, O any] func(ctx context.Context, in *I) (*O, error)

// AuthedHandler is a Handler that also receives the validated access token data.
type AuthedHandler[I, O, U any] func(ctx context.Context, in *I, user *U) (*O, error)

// Handle registers an Echo route and its OpenAPI operation. The input and
// output types are reflected into the OpenAPI document, while the handler is
// wrapped in a plain Echo handler so it remains compatible with Echo
// middleware
func (api *Api) Handle[I, O any](
	method string,
	path string,
	route Route,
	handler Handler[I, O],
	middleware ...echo.MiddlewareFunc,
) {
	api.registerOperation(method, path, route, reflect.TypeFor[I](), reflect.TypeFor[O]())
	api.Engine.Add(method, path, api.wrap(route, handler), middleware...)
}

// HandleAuthed registers an Echo route whose handler also receives the access
// token data validated by the authentication middleware.
func (api *Api) HandleAuthed[I, O, U any](
	method string,
	path string,
	route Route,
	handler AuthedHandler[I, O, U],
	middleware ...echo.MiddlewareFunc,
) {
	route.Auth = true
	api.registerOperation(method, path, route, reflect.TypeFor[I](), reflect.TypeFor[O]())
	api.Engine.Add(method, path, api.wrapAuthed(route, handler), middleware...)
}

func (api *Api) wrap[I, O any](route Route, handler Handler[I, O]) echo.HandlerFunc {
	return func(c *echo.Context) error {
		var input I
		if err := bindAndValidate(c, &input); err != nil {
			return err
		}

		output, err := handler(c.Request().Context(), &input)
		if err != nil {
			return err
		}

		return api.writeResponse(c, route, output)
	}
}

func (api *Api) wrapAuthed[I, O, U any](route Route, handler AuthedHandler[I, O, U]) echo.HandlerFunc {
	return func(c *echo.Context) error {
		var input I
		if err := bindAndValidate(c, &input); err != nil {
			return err
		}

		var user U
		if !injectUserAccess(c.Request().Context(), &user) {
			return apperrors.InvalidToken
		}

		output, err := handler(c.Request().Context(), &input, &user)
		if err != nil {
			return err
		}

		return api.writeResponse(c, route, output)
	}
}

func bindAndValidate(c *echo.Context, input any) error {
	if err := c.Bind(input); err != nil {
		return err
	}
	if err := echo.BindHeaders(c, input); err != nil {
		return err
	}
	return c.Validate(input)
}

func injectUserAccess(ctx context.Context, target any) bool {
	user, ok := UserAccessFromContext(ctx)
	if !ok {
		return false
	}

	userValue := reflect.ValueOf(user)
	if !userValue.IsValid() {
		return false
	}

	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return false
	}
	value = value.Elem()

	if !userValue.Type().AssignableTo(value.Type()) {
		return false
	}
	value.Set(userValue)
	return true
}

func (api *Api) writeResponse(c *echo.Context, route Route, output any) error {
	status := route.Status
	if status == 0 {
		status = http.StatusOK
	}
	if output == nil {
		return c.NoContent(status)
	}

	value := reflect.ValueOf(output)
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return c.NoContent(status)
		}
		value = value.Elem()
	}

	if strings.HasPrefix(route.ContentType, "text/") {
		return c.String(status, fmt.Sprint(value.Interface()))
	}

	return c.JSON(status, output)
}

func (api *Api) registerOperation(method, path string, route Route, input, output reflect.Type) {
	if api.OpenAPI == nil {
		api.OpenAPI = newOpenAPIDocument()
	}

	operation := &openapiOperation{
		ID:          route.ID,
		Summary:     route.Summary,
		Description: route.Description,
		Tag:         route.Tag,
		Path:        path,
		ContentType: route.ContentType,
		Errors:      route.Errors,
		Auth:        route.Auth,
		Permissions: route.Permissions,
		Status:      route.Status,
		Input:       input,
		Output:      output,
		Schemas:     api.OpenAPI.Components.Schemas,
	}
	api.OpenAPI.AddOperation(toOpenAPIPath(path), method, operation.build())
}
