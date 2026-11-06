package users_test

import (
	"encoding/json"
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
	users.Register(api, users.Deps{Middleware: network.NewMiddleware(nil, nil)})

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

	var document struct {
		Paths map[string]struct {
			Get struct {
				RequestBody json.RawMessage `json:"requestBody"`
			} `json:"get"`
			Post struct {
				RequestBody json.RawMessage `json:"requestBody"`
			} `json:"post"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &document); err != nil {
		t.Fatalf("decode OpenAPI document: %v", err)
	}
	if len(document.Paths["/api/v1/users/me"].Get.RequestBody) != 0 {
		t.Fatal("token-only route /api/v1/users/me should not document a request body")
	}
	if len(document.Paths["/api/v1/users/mfa/setup"].Post.RequestBody) != 0 {
		t.Fatal("token-only route /api/v1/users/mfa/setup should not document a request body")
	}
}
