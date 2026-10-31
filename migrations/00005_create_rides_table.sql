-- +goose Up
CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TYPE ride_status AS ENUM (
    'requested',
    'driver_assigned',
    'driver_arrived',
    'in_progress',
    'completed',
    'cancelled'
);

CREATE TABLE IF NOT EXISTS rides (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    customer_id BIGINT NOT NULL REFERENCES users(id),
    driver_id BIGINT REFERENCES users(id) DEFAULT NULL,

    origin GEOGRAPHY(POINT, 4326) NOT NULL,
    destination GEOGRAPHY(POINT, 4326) NOT NULL,
    
    status ride_status NOT NULL DEFAULT 'requested',
    fare INT DEFAULT NULL CHECK (fare >= 0),
    
    accepted_at TIMESTAMPTZ DEFAULT NULL,
    started_at TIMESTAMPTZ DEFAULT NULL,
    completed_at TIMESTAMPTZ DEFAULT NULL,
    
    cancelled_at TIMESTAMPTZ DEFAULT NULL,
    cancelled_by BIGINT REFERENCES users(id) DEFAULT NULL,
    cancellation_reason TEXT DEFAULT NULL,
	estimated_time_of_arrival TIMESTAMPTZ,
    
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT check_cancellation_user CHECK (
        cancelled_by IS NULL OR cancelled_by IN (customer_id, driver_id)
    )
);

CREATE INDEX idx_rides_customer_id ON rides(customer_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_rides_driver_id ON rides(driver_id) WHERE deleted_at IS NULL AND driver_id IS NOT NULL;
CREATE INDEX idx_rides_status ON rides(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_rides_origin_gist ON rides USING GIST(origin);

CREATE TRIGGER rides_updated_at
BEFORE UPDATE ON rides
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

INSERT INTO permissions (name)
VALUES
    ('ride.read'),
    ('ride.create'),
    ('ride.cancel'),
    ('ride.delete'),
    ('ride.accept'),
    ('ride.arrive'),
    ('ride.start'),
    ('ride.complete')
ON CONFLICT (name) DO NOTHING;

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN (
    'ride.read',
    'ride.create',
    'ride.cancel',
    'ride.delete'
)
WHERE r.name = 'customer'
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN (
    'ride.read',
    'ride.cancel',
    'ride.accept',
    'ride.arrive',
    'ride.start',
    'ride.complete'
)
WHERE r.name = 'driver'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- +goose Down
DELETE FROM permissions
WHERE name IN (
    'ride.read',
    'ride.create',
    'ride.cancel',
    'ride.delete',
    'ride.accept',
    'ride.arrive',
    'ride.start',
    'ride.complete'
);

DROP TABLE IF EXISTS rides;
