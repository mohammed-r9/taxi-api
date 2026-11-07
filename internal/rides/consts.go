package rides

import "time"

const (
	rideQouteKeyPrefix = "ride_qoute:user:"  // + user id
	userHasRidePrefix  = "user:active_ride:" // + user id
	rideQouteExpiry    = time.Minute * 2
)
