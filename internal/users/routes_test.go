package users_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/mohammed-r9/taxi-api/internal/network"
	"github.com/mohammed-r9/taxi-api/internal/users"
)

func TestUsersRoutesAreRegistered(t *testing.T) {
	e := echo.New()
	api := network.NewApi(e, network.NewRequestValidator())
	api.Init()
	users.Register(api, users.Deps{})

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/docs/openapi.json", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}

	spec := recorder.Body.String()
	for _, path := range []string{
		`"/api/v1/users/register"`,
		`"/api/v1/users/login"`,
		`"/api/v1/users/me"`,
		`"/api/v1/users/mfa/setup"`,
		`"/api/v1/users/mfa/enable"`,
		`"/api/v1/users/mfa/verify"`,
	} {
		if !strings.Contains(spec, path) {
			t.Errorf("openapi document does not contain %s", path)
		}
	}
}
