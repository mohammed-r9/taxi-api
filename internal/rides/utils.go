package rides

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/mohammed-r9/taxi-api/internal/apperrors"
	"github.com/redis/go-redis/v9"
)

func rideQouteKey(userID int64) string {
	return fmt.Sprintf("%v%v", rideQouteKeyPrefix, userID)
}

func userActiveRideKey(userID int64) string {
	return fmt.Sprintf("%v%v", userHasRidePrefix, userID)
}

func hasActiveRide(ctx context.Context, rdb *redis.Client, userID int64) (bool, error) {
	exists, err := rdb.Exists(ctx, userActiveRideKey(userID)).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func missingRideQouteError(ctx context.Context, rdb *redis.Client, userID int64) error {
	hasRide, err := hasActiveRide(ctx, rdb, userID)
	if err != nil {
		return err
	}
	if hasRide {
		return apperrors.UserAlreadyOnRide
	}
	return apperrors.RideNotFound
}

func newRideQuote(ctx context.Context, userID int64, origin, destination Location) (RideQoute, error) {
	fare, err := CalculateFare(ctx, origin, destination)
	if err != nil {
		return RideQoute{}, err
	}

	now := time.Now().UTC()

	return RideQoute{
		UserID:      userID,
		Origin:      origin,
		Destination: destination,
		Fare:        fare,
		QouteID:     uuid.NewV4().String(),
		CreatedAt:   now,
		ETA:         now.Add(30 * time.Minute), // temp
	}, nil
}
