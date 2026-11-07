package apperrors

import "net/http"

const (
	RideNotFound rideError = iota
	RideQuoteNotFound
	InvalidRideID
	InvalidRideStatus
	RideAlreadyAccepted
	RideAlreadyCancelled
	RideAlreadyCompleted
	RideNotAssigned
	CannotCancelRide
	UnauthorizedRideAccess
	InvalidLocation
	InvalidFare
	DriverAlreadyOnRide
	NoDriversAvailable
	UserAlreadyOnRide

	// new errors should be added above this line
	rideErrorCount
)

var rideErrorTable = [rideErrorCount]errorInfo{
	RideNotFound: {
		message: "ride not found",
		code:    "RIDE_NOT_FOUND",
		status:  http.StatusNotFound,
	},
	RideQuoteNotFound: {
		message: "ride quote not found or expired",
		code:    "RIDE_QUOTE_NOT_FOUND",
		status:  http.StatusNotFound,
	},
	InvalidRideID: {
		message: "invalid ride id",
		code:    "INVALID_RIDE_ID",
		status:  http.StatusBadRequest,
	},
	InvalidRideStatus: {
		message: "ride is not in a valid state for this action",
		code:    "INVALID_RIDE_STATUS",
		status:  http.StatusConflict,
	},
	RideAlreadyAccepted: {
		message: "ride has already been accepted",
		code:    "RIDE_ALREADY_ACCEPTED",
		status:  http.StatusConflict,
	},
	RideAlreadyCancelled: {
		message: "ride has already been cancelled",
		code:    "RIDE_ALREADY_CANCELLED",
		status:  http.StatusConflict,
	},
	RideAlreadyCompleted: {
		message: "ride has already been completed",
		code:    "RIDE_ALREADY_COMPLETED",
		status:  http.StatusConflict,
	},
	RideNotAssigned: {
		message: "ride has no assigned driver",
		code:    "RIDE_NOT_ASSIGNED",
		status:  http.StatusConflict,
	},
	CannotCancelRide: {
		message: "ride cannot be cancelled in its current state",
		code:    "CANNOT_CANCEL_RIDE",
		status:  http.StatusConflict,
	},
	UnauthorizedRideAccess: {
		message: "not authorized to access this ride",
		code:    "UNAUTHORIZED_RIDE_ACCESS",
		status:  http.StatusForbidden,
	},
	InvalidLocation: {
		message: "invalid location",
		code:    "INVALID_LOCATION",
		status:  http.StatusBadRequest,
	},
	InvalidFare: {
		message: "invalid fare",
		code:    "INVALID_FARE",
		status:  http.StatusBadRequest,
	},
	DriverAlreadyOnRide: {
		message: "driver already has an active ride",
		code:    "DRIVER_ALREADY_ON_RIDE",
		status:  http.StatusConflict,
	},
	NoDriversAvailable: {
		message: "no drivers available",
		code:    "NO_DRIVERS_AVAILABLE",
		status:  http.StatusServiceUnavailable,
	},
	UserAlreadyOnRide: {
		message: "user has an active ride already",
		code:    "USER_ALREADY_ON_RIDE",
		status:  http.StatusConflict,
	},
}

var unhandledRideError = errorInfo{
	message: "unhandled ride error",
	code:    "UNHANDLED_RIDE_ERROR",
	status:  http.StatusInternalServerError,
}

func (e rideError) Error() string {
	return e.info().message
}

func (e rideError) Code() string {
	return e.info().code
}

func (e rideError) Status() int {
	return e.info().status
}

func (e rideError) info() errorInfo {
	if int(e) < 0 || int(e) >= len(rideErrorTable) {
		return unhandledRideError
	}

	info := rideErrorTable[e]

	if info.message == "" {
		return unhandledRideError
	}

	return info
}
