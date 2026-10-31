# go-boiler

A starting point for new Go HTTP APIs: project layout, dependency wiring,
request validation, error envelopes, structured logging, JWT/password/TOTP
helpers, Postgres + Redis connections, sqlc and goose setup.

## Layout

```
cmd/                 application entrypoint and lifecycle (startup, shutdown)
db/                  Postgres pool, Redis client, Store (transaction wrapper)
internal/apperrors/  typed application errors with HTTP status + code
internal/auth/       JWT, Argon2id passwords, opaque tokens, TOTP, recovery codes
internal/network/    Huma + Echo wiring, routing, validation, logging, rate limits
migrations/          goose migrations
queries/             sqlc queries (generated code goes to db/)
```

## Getting started

1. Copy this directory and rename the module in `go.mod`:
   ```sh
   sed -i 's#github.com/mohammed-r9/go-boiler#github.com/you/your-project#g' go.mod $(find . -name '*.go')
   ```
2. `cp .env.example .env` and fill in the values.
3. Start dependencies: `docker compose up -d`
4. Run migrations: `just db-up`
5. Start the API: `just dev`

Docs are served at `http://localhost:8080/api/docs`; the OpenAPI document is at
`/api/docs/openapi.json`.

## Adding an endpoint

Each feature lives in its own package under `internal/` and exposes a
`Register(api *network.Api)` function. Handler implementations stay in the
feature package; `network.Handle` only wires them to the router and the docs.

```go
// internal/things/handler.go
package things

type Handler struct{ validate *network.RequestValidator }

func New(v *network.RequestValidator) *Handler { return &Handler{validate: v} }

type CreateInput struct {
	Body struct {
		Name string `json:"name" validate:"required,min=3" minLength:"3"`
	}
}

type CreateOutput struct {
	Body network.BaseResponse[network.Map]
}

func (h *Handler) Create(ctx context.Context, input *CreateInput) (*CreateOutput, error) {
	if err := h.validate.Validate(&input.Body); err != nil {
		return nil, err
	}
	return &CreateOutput{Body: network.BaseResponse[network.Map]{
		Success: true, Status: http.StatusCreated, Message: "created",
		Data: network.Map{"name": input.Body.Name},
	}}, nil
}
```

```go
// internal/things/routes.go
package things

func Register(api *network.Api, h *Handler) {
	network.Handle(api, http.MethodPost, "/api/v1/things", network.Route{
		ID: "createThing", Summary: "Create a thing", Tag: "Things",
		Status: http.StatusCreated, Auth: true,
	}, h.Create)
}
```

Then register the feature from the composition root in
`internal/router/router.go`:

```go
func Register(api *network.Api, deps Deps) {
	api.Init()

	health.Register(api)
	users.Register(api)
	things.Register(api, things.New(deps.Validate))
}
```

## Adding database code

1. Add a migration: `just db-create <name>` then edit it in `migrations/`.
2. Add queries to `queries/*.sql`.
3. Run `just generate` (sqlc writes to `db/`).
4. Embed `*Queries` on `db.Store` so handlers can call generated methods.

## Commands

| Command | Description |
| --- | --- |
| `just dev` | Run the API |
| `just live` | Run with restart on file changes |
| `just build` | Build to `bin/app` |
| `just test` | Run all tests |
| `just generate` | Run sqlc |
| `just db-up` | Apply migrations |
| `just db-down` | Roll back one migration |
| `just db-status` | Show migration status |
