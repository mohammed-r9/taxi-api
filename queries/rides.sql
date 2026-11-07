-- name: CreateRide :one
INSERT INTO rides (
	customer_id,
	origin,
	destination,
	fare,
	estimated_time_of_arrival
) VALUES (
	@customer_id,
	ST_SetSRID(ST_MakePoint(@origin_longitude::double precision, @origin_latitude::double precision), 4326)::geography,
	ST_SetSRID(ST_MakePoint(@destination_longitude::double precision, @destination_latitude::double precision), 4326)::geography,
	COALESCE(@fare::double precision, 0),
	COALESCE(@estimated_time_of_arrival::timestamptz, CURRENT_TIMESTAMP)
)
RETURNING id;
