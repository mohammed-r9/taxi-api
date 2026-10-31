package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/mohammed-r9/taxi-api/internal/network"
)

type middlewareOutput struct {
	Body string `contentType:"text/plain"`
}

func TestRequireAuthPopulatesHandlerContext(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	e := echo.New()
	api := network.NewApi(e, network.NewRequestValidator())
	api.Init()

	token, err := GenerateAccessToken(UserAccessData{UserID: 42, RoleID: 2})
	if err != nil {
		t.Fatal(err)
	}

	network.Handle(api, http.MethodGet, "/protected", network.Route{
		ID: "protected",
	}, func(ctx context.Context, _ *struct{}) (*middlewareOutput, error) {
		user, ok := UserFromContext(ctx)
		if !ok {
			return nil, errors.New("user missing from context")
		}
		if user.UserID != 42 || user.RoleID != 2 {
			return nil, errors.New("unexpected user")
		}
		return &middlewareOutput{Body: "ok"}, nil
	}, RequireAuth)

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestRequireAuthRejectsMissingHeader(t *testing.T) {
	e := echo.New()
	api := network.NewApi(e, network.NewRequestValidator())
	api.Init()

	network.Handle(api, http.MethodGet, "/protected", network.Route{
		ID: "protectedMissingHeader",
	}, func(context.Context, *struct{}) (*middlewareOutput, error) {
		return &middlewareOutput{Body: "ok"}, nil
	}, RequireAuth)

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/protected", nil))

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
}
