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

type testOutput string

func newTestAPI() *echo.Echo {
	e := echo.New()
	api := network.NewApi(e, network.NewRequestValidator())
	api.Init()
	health.Register(api)
	return e
}

func TestHealth(t *testing.T) {
	e := newTestAPI()
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "status is available") {
		t.Fatalf("body = %q", recorder.Body.String())
	}
}

func TestOpenAPIDocumentation(t *testing.T) {
	e := newTestAPI()

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
	e := newTestAPI()
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

type echoRequest struct {
	Name string `json:"name" validate:"required"`
}

type echoResponse struct {
	Greeting string `json:"greeting"`
}

func TestHandleBindsInputAndWritesOutput(t *testing.T) {
	e := echo.New()
	api := network.NewApi(e, network.NewRequestValidator())
	api.Init()

	api.Handle(http.MethodPost, "/echo", network.Route{
		ID: "echo", Status: http.StatusCreated,
	}, func(_ context.Context, in *echoRequest) (*echoResponse, error) {
		return &echoResponse{Greeting: "hello " + in.Name}, nil
	})

	request := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(`{"name":"world"}`))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if strings.TrimSpace(recorder.Body.String()) != `{"greeting":"hello world"}` {
		t.Fatalf("body = %q", recorder.Body.String())
	}
}

type pathInput struct {
	ID int64 `param:"id"`
}

type pathOutput struct {
	ID int64 `json:"id"`
}

func TestHandleBindsPathParameters(t *testing.T) {
	e := echo.New()
	api := network.NewApi(e, network.NewRequestValidator())
	api.Init()

	api.Handle(http.MethodGet, "/api/v1/rides/:id", network.Route{
		ID: "getRide",
	}, func(_ context.Context, in *pathInput) (*pathOutput, error) {
		return &pathOutput{ID: in.ID}, nil
	})

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/rides/42", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if strings.TrimSpace(recorder.Body.String()) != `{"id":42}` {
		t.Fatalf("body = %q", recorder.Body.String())
	}

	specRecorder := httptest.NewRecorder()
	e.ServeHTTP(specRecorder, httptest.NewRequest(http.MethodGet, "/api/docs/openapi.json", nil))
	if !strings.Contains(specRecorder.Body.String(), `"/api/v1/rides/{id}"`) {
		t.Fatalf("OpenAPI document does not contain the path parameter route")
	}
}

type queryHeaderInput struct {
	Term string `query:"term"`
	Key  string `header:"X-Api-Key"`
}

type queryHeaderOutput struct {
	Term string `json:"term"`
	Key  string `json:"key"`
}

func TestHandleBindsQueryAndHeaders(t *testing.T) {
	e := echo.New()
	api := network.NewApi(e, network.NewRequestValidator())
	api.Init()

	api.Handle(http.MethodGet, "/search", network.Route{
		ID: "search",
	}, func(_ context.Context, in *queryHeaderInput) (*queryHeaderOutput, error) {
		return &queryHeaderOutput{Term: in.Term, Key: in.Key}, nil
	})

	request := httptest.NewRequest(http.MethodGet, "/search?term=ride", nil)
	request.Header.Set("X-Api-Key", "secret")
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if strings.TrimSpace(recorder.Body.String()) != `{"term":"ride","key":"secret"}` {
		t.Fatalf("body = %q", recorder.Body.String())
	}
}

type mixedInput struct {
	ID    int64  `param:"id" json:"id" validate:"required"`
	Limit int    `query:"limit" json:"limit"`
	Key   string `header:"X-Api-Key" json:"key"`
	Name  string `json:"name" validate:"required"`
}

type mixedOutput struct {
	OK bool `json:"ok"`
}

func TestOpenAPIDocumentsBodyFieldsSeparatelyFromParameters(t *testing.T) {
	e := echo.New()
	api := network.NewApi(e, network.NewRequestValidator())
	api.Init()

	api.Handle(http.MethodPost, "/api/v1/mixed/:id", network.Route{
		ID: "mixed",
	}, func(context.Context, *mixedInput) (*mixedOutput, error) {
		return &mixedOutput{OK: true}, nil
	})

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/docs/openapi.json", nil))

	var spec struct {
		Paths map[string]struct {
			Post struct {
				Parameters []struct {
					Name string `json:"name"`
					In   string `json:"in"`
				} `json:"parameters"`
				RequestBody struct {
					Content map[string]struct {
						Schema struct {
							Properties map[string]json.RawMessage `json:"properties"`
							Required   []string                   `json:"required"`
						} `json:"schema"`
					} `json:"content"`
				} `json:"requestBody"`
			} `json:"post"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &spec); err != nil {
		t.Fatalf("decode OpenAPI document: %v", err)
	}

	post := spec.Paths["/api/v1/mixed/{id}"].Post
	parameters := make(map[string]string)
	for _, parameter := range post.Parameters {
		parameters[parameter.Name] = parameter.In
	}
	if parameters["id"] != "path" || parameters["limit"] != "query" || parameters["X-Api-Key"] != "header" {
		t.Fatalf("parameters = %+v", parameters)
	}

	schema := post.RequestBody.Content["application/json"].Schema
	if len(schema.Properties) != 1 {
		t.Fatalf("body properties = %+v", schema.Properties)
	}
	if _, ok := schema.Properties["name"]; !ok {
		t.Fatalf("body properties = %+v", schema.Properties)
	}
	if len(schema.Required) != 1 || schema.Required[0] != "name" {
		t.Fatalf("body required = %+v", schema.Required)
	}
}

type injectedAccess struct {
	UserID int64 `validate:"max=10"`
}

type injectedOutput struct {
	UserID int64 `json:"user_id"`
}

func TestHandleInjectsUserAccessAfterValidation(t *testing.T) {
	e := echo.New()
	api := network.NewApi(e, network.NewRequestValidator())
	api.Init()

	middleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx := network.WithUserAccess(c.Request().Context(), injectedAccess{UserID: 42})
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
	api.HandleAuthed(http.MethodGet, "/injected", network.Route{
		ID: "injected",
	}, func(_ context.Context, _ *struct{}, in *injectedAccess) (*injectedOutput, error) {
		return &injectedOutput{UserID: in.UserID}, nil
	}, middleware)

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/injected", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if strings.TrimSpace(recorder.Body.String()) != `{"user_id":42}` {
		t.Fatalf("body = %q", recorder.Body.String())
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
	api.Handle(http.MethodGet, "/api/v1/middleware-test", network.Route{
		ID: "middlewareTest", Summary: "Test middleware", Tag: "System",
	}, func(context.Context, *struct{}) (*testOutput, error) {
		output := testOutput("ok")
		return &output, nil
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

func TestOpenAPIDocumentsUnifiedErrorResponseForAllHandlers(t *testing.T) {
	e := echo.New()
	api := network.NewApi(e, network.NewRequestValidator())
	api.Init()

	for _, path := range []string{"/error-one", "/error-two"} {
		api.Handle(http.MethodGet, path, network.Route{
			ID:     strings.TrimPrefix(path, "/"),
			Errors: []int{http.StatusBadRequest},
		}, func(context.Context, *struct{}) (*testOutput, error) {
			output := testOutput("ok")
			return &output, nil
		})
	}

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/docs/openapi.json", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}

	var spec struct {
		Paths map[string]struct {
			Get struct {
				Responses map[string]struct {
					Content map[string]struct {
						Schema struct {
							Ref string `json:"$ref"`
						} `json:"schema"`
						Example struct {
							Success bool   `json:"success"`
							Status  int    `json:"status"`
							Code    string `json:"code"`
							Message string `json:"message"`
						} `json:"example"`
					} `json:"content"`
				} `json:"responses"`
			} `json:"get"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &spec); err != nil {
		t.Fatalf("decode OpenAPI document: %v", err)
	}
	for _, path := range []string{"/error-one", "/error-two"} {
		for _, code := range []string{"default", "400"} {
			response, ok := spec.Paths[path].Get.Responses[code]
			if !ok {
				t.Fatalf("%s response is not documented for %s", code, path)
			}

			media, ok := response.Content["application/json"]
			if !ok {
				t.Fatalf("%s response has no JSON content for %s", code, path)
			}
			if media.Schema.Ref != "#/components/schemas/ErrorResponse" {
				t.Fatalf("%s response schema = %q", code, media.Schema.Ref)
			}
			if media.Example.Success || media.Example.Status != http.StatusBadRequest ||
				media.Example.Code != "VALIDATION_ERROR" || media.Example.Message != "request validation failed" {
				t.Fatalf("%s response example = %+v", code, media.Example)
			}
		}
	}
}
