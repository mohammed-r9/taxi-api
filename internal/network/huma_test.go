package network_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/mohammed-r9/taxi-api/internal/health"
	"github.com/mohammed-r9/taxi-api/internal/network"
)

type testOutput struct {
	Body string `contentType:"text/plain"`
}

func newHumaTestAPI() *echo.Echo {
	e := echo.New()
	api := network.NewApi(e, network.NewRequestValidator())
	api.Init()
	health.Register(api)
	return e
}

func TestHealth(t *testing.T) {
	e := newHumaTestAPI()
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "status is available") {
		t.Fatalf("body = %q", recorder.Body.String())
	}
}

func TestHumaDocumentation(t *testing.T) {
	e := newHumaTestAPI()

	for _, test := range []struct {
		path        string
		contentType string
		contains    string
	}{
		{path: "/api/docs", contentType: "text/html", contains: "@scalar/api-reference"},
		{path: "/api/docs/openapi.json", contentType: "application/openapi+json", contains: `"/api/v1/health"`},
	} {
		recorder := httptest.NewRecorder()
		e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
		if recorder.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want 200", test.path, recorder.Code)
		}
		if !strings.Contains(recorder.Header().Get("Content-Type"), test.contentType) {
			t.Errorf("GET %s Content-Type = %q", test.path, recorder.Header().Get("Content-Type"))
		}
		if !strings.Contains(recorder.Body.String(), test.contains) {
			t.Errorf("GET %s response does not contain %q", test.path, test.contains)
		}
	}
}

func TestHandleDocumentsBearerAuth(t *testing.T) {
	e := newHumaTestAPI()
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/docs/openapi.json", nil))

	var spec struct {
		Components struct {
			SecuritySchemes map[string]struct {
				Type   string `json:"type"`
				Scheme string `json:"scheme"`
			} `json:"securitySchemes"`
		} `json:"components"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &spec); err != nil {
		t.Fatalf("decode OpenAPI document: %v", err)
	}
	bearer := spec.Components.SecuritySchemes["BearerAuth"]
	if bearer.Type != "http" || bearer.Scheme != "bearer" {
		t.Fatalf("BearerAuth = %+v", bearer)
	}
}

func TestHandleAcceptsEchoMiddlewareDirectly(t *testing.T) {
	e := echo.New()
	api := network.NewApi(e, network.NewRequestValidator())
	api.Init()

	middleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Response().Header().Set("X-Test-Middleware", "ran")
			return next(c)
		}
	}
	network.Handle(api, http.MethodGet, "/api/v1/middleware-test", network.Route{
		ID: "middlewareTest", Summary: "Test middleware", Tag: "System",
	}, func(context.Context, *struct{}) (*testOutput, error) {
		return &testOutput{Body: "ok"}, nil
	}, middleware)

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/middleware-test", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if recorder.Header().Get("X-Test-Middleware") != "ran" {
		t.Fatal("Echo middleware was not run")
	}
}
