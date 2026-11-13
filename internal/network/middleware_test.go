package network

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/mohammed-r9/taxi-api/internal/auth"
	"github.com/stretchr/testify/require"
)

type authTestOutput string

func TestAuthPopulatesHandlerContext(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	e := echo.New()
	api := NewApi(e, NewRequestValidator())
	api.Init()

	token, err := auth.GenerateAccessToken(auth.UserAccessData{UserID: 42, RoleID: 2})
	require.NoError(t, err)

	api.Handle(http.MethodGet, "/protected", Route{
		ID: "protected",
	}, func(ctx context.Context, _ *struct{}) (*authTestOutput, error) {
		user, ok := UserAccessFromContext(ctx)
		if !ok {
			return nil, errors.New("user missing from context")
		}
		data, ok := user.(auth.UserAccessData)
		if !ok || data.UserID != 42 || data.RoleID != 2 {
			return nil, errors.New("unexpected user")
		}
		output := authTestOutput("ok")
		return &output, nil
	}, NewMiddleware(nil, nil).Auth())

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
}

func TestHandleInjectsUserAccess(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	e := echo.New()
	api := NewApi(e, NewRequestValidator())
	api.Init()

	token, err := auth.GenerateAccessToken(auth.UserAccessData{UserID: 42, RoleID: 2})
	require.NoError(t, err)

	api.HandleAuthed(http.MethodGet, "/protected-injected", Route{
		ID: "protectedInjected",
	}, func(_ context.Context, _ *struct{}, in *auth.UserAccessData) (*authTestOutput, error) {
		if in.UserID != 42 || in.RoleID != 2 {
			return nil, errors.New("unexpected user access")
		}
		output := authTestOutput("ok")
		return &output, nil
	}, NewMiddleware(nil, nil).Auth())

	request := httptest.NewRequest(http.MethodGet, "/protected-injected", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
}

func TestAuthRejectsMissingHeader(t *testing.T) {
	e := echo.New()
	api := NewApi(e, NewRequestValidator())
	api.Init()

	api.Handle(http.MethodGet, "/protected", Route{
		ID: "protectedMissingHeader",
	}, func(context.Context, *struct{}) (*authTestOutput, error) {
		output := authTestOutput("ok")
		return &output, nil
	}, NewMiddleware(nil, nil).Auth())

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/protected", nil))

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}
