package health

import (
	"net/http"

	"github.com/mohammed-r9/taxi-api/internal/network"
)

func Register(api *network.Api) {
	h := New()

	network.Handle(api, http.MethodGet, "/api/v1/health", network.Route{
		ID:      "health",
		Summary: "Check API availability",
		Tag:     "System",
	}, h.Check)
}
