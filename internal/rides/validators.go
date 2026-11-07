package rides

import "math"

func validLocation(l Location) bool {
	return (!math.IsNaN(l.Latitude) &&
		!math.IsInf(l.Latitude, 0) &&
		l.Latitude >= -90 &&
		l.Latitude <= 90) && (!math.IsNaN(l.Longitude) &&
		!math.IsInf(l.Longitude, 0) &&
		l.Longitude >= -180 &&
		l.Longitude <= 180)
}
