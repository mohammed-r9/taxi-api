package rides

import (
	"time"

	"github.com/mohammed-r9/taxi-api/internal/network"
)

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type RideQoute struct {
	UserID      int64     `json:"user_id"`
	QouteID     string    `json:"qoute_id"`
	Origin      Location  `json:"origin"`
	Destination Location  `json:"destination"`
	Fare        float64   `json:"fare"`
	CreatedAt   time.Time `json:"created_at"`
	ETA         time.Time `json:"eta"`
}

type RequestRideInput struct {
	Origin      Location `json:"origin"`
	Destination Location `json:"destination"`
}

type ConfirmRideInput struct {
	ID string `param:"id" validate:"required,uuid" format:"uuid"`
}

type ConfirmRideResponse struct {
	RideID int64     `json:"ride_id"`
	Status string    `json:"status"`
	Fare   float64   `json:"fare"`
	ETA    time.Time `json:"eta"`
}

type ConfirmRideOutput = network.ResponseOutput[ConfirmRideResponse]

type RequestRideResponse struct {
	QouteID string    `json:"qoute_id"`
	Fare    float64   `json:"fare"`
	ETA     time.Time `json:"eta"`
}

type RequestRideOutput = network.ResponseOutput[RequestRideResponse]
