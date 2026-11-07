package rides

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mohammed-r9/taxi-api/db"
	"github.com/mohammed-r9/taxi-api/internal/apperrors"
	"github.com/mohammed-r9/taxi-api/internal/auth"
	"github.com/mohammed-r9/taxi-api/internal/network"
	"github.com/redis/go-redis/v9"
)

type Deps struct {
	Store      *db.Store
	Redis      *redis.Client
	Validator  *network.RequestValidator
	Middleware *network.Middleware
}

type Handler struct {
	Deps
}

func New(deps Deps) *Handler {
	return &Handler{Deps: deps}
}

func (h *Handler) RequestRide(ctx context.Context, in *RequestRideInput, user *auth.UserAccessData) (*RequestRideOutput, error) {
	isLocationValid := validLocation(in.Destination) && validLocation(in.Origin)
	if !isLocationValid {
		return nil, apperrors.InvalidLocation
	}

	hasRide, err := hasActiveRide(ctx, h.Redis, user.UserID)
	if err != nil {
		return nil, err
	}

	if hasRide {
		return nil, apperrors.UserAlreadyOnRide
	}

	ride, err := newRideQuote(ctx, user.UserID, in.Origin, in.Destination)
	if err != nil {
		return nil, err
	}

	rideJson, err := json.Marshal(ride)
	if err != nil {
		return nil, err
	}
	err = h.Redis.Set(ctx, rideQouteKey(user.UserID), rideJson, rideQouteExpiry).Err()
	if err != nil {
		return nil, err
	}

	return network.NewSuccessResponse(http.StatusOK, RequestRideResponse{
		QouteID: ride.QouteID,
		Fare:    ride.Fare,
		ETA:     ride.ETA,
	}, "ride initialized successfully"), nil
}

func (h *Handler) ConfirmRide(ctx context.Context, in *ConfirmRideInput, user *auth.UserAccessData) (*ConfirmRideOutput, error) {
	rideQouteJson, err := h.Redis.GetDel(ctx, rideQouteKey(user.UserID)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, missingRideQouteError(ctx, h.Redis, user.UserID)
	}
	if err != nil {
		return nil, err
	}

	ride := RideQoute{}
	err = json.Unmarshal([]byte(rideQouteJson), &ride)
	if err != nil {
		return nil, err
	}
	if ride.UserID != user.UserID {
		return nil, apperrors.UnauthorizedRideAccess
	}

	if ride.QouteID != in.ID {
		return nil, apperrors.RideNotFound
	}

	rideID, err := h.Store.Queries.CreateRide(ctx, db.CreateRideParams{
		CustomerID:             user.UserID,
		Fare:                   ride.Fare,
		OriginLatitude:         ride.Origin.Latitude,
		OriginLongitude:        ride.Origin.Longitude,
		DestinationLatitude:    ride.Destination.Latitude,
		DestinationLongitude:   ride.Destination.Longitude,
		EstimatedTimeOfArrival: ride.ETA,
	})
	if err != nil {
		return nil, err
	}

	err = h.Redis.Set(ctx, userActiveRideKey(user.UserID), rideID, 0).Err()
	if err != nil {
		return nil, err
	}

	return network.NewSuccessResponse(http.StatusOK, ConfirmRideResponse{
		RideID: rideID,
		Status: "requested",
		Fare:   ride.Fare,
		ETA:    ride.ETA,
	}, "ride confirmed successfully"), nil
}
