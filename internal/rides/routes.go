package rides

import (
	"net/http"
	"time"

	"github.com/mohammed-r9/taxi-api/internal/network"
)

func Register(api *network.Api, deps Deps) {
	handler := New(deps)
	mw := deps.Middleware

	api.HandleAuthed(http.MethodPost, "/api/v1/rides/request", network.Route{
		ID:      "RideRequest",
		Summary: "Request A Ride",
		Tag:     "Rides",
		Status:  http.StatusOK,
	}, handler.RequestRide,
		mw.Auth("ride.create"),
		mw.RateLimit(network.Policy{
			Attempts: 5,
			Window:   2 * time.Minute,
		}, network.RateLimitByUserID))

	api.HandleAuthed(http.MethodPost, "/api/v1/api/rides/:id", network.Route{
		ID:      "RideConfirm",
		Summary: "Confirm A Ride",
		Tag:     "Rides",
		Status:  http.StatusOK,
	}, handler.ConfirmRide,
		mw.Auth("ride.create"))
}
