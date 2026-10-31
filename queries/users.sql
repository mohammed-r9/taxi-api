-- name: CreateUser :one
INSERT INTO users (email, name, username, password, role_id)
VALUES (@email, @name, @username, @password, @role_id)
RETURNING id, email, name, username, role_id, two_fa_enabled, verified_at, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id,
       email,
       name,
       username,
       password,
       role_id,
       two_fa_secret,
       two_fa_enabled,
       two_fa_encryption_version,
       verified_at,
       created_at,
       updated_at
FROM users
WHERE email = @email
  AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT id,
       email,
       name,
       username,
       password,
       role_id,
       two_fa_secret,
       two_fa_enabled,
       two_fa_encryption_version,
       verified_at,
       created_at,
       updated_at
FROM users
WHERE id = @id
  AND deleted_at IS NULL;

-- name: SetUserTOTPSecret :exec
UPDATE users
SET two_fa_secret             = @two_fa_secret,
    two_fa_encryption_version = @two_fa_encryption_version,
    two_fa_enabled            = FALSE
WHERE id = @id;

-- name: EnableUserTOTP :exec
UPDATE users
SET two_fa_enabled = TRUE
WHERE id = @id;

-- name: CreateSession :exec
INSERT INTO sessions (user_id, refresh_token, device_name, expires_at)
VALUES (@user_id, @refresh_token, @device_name, @expires_at);

-- name: GetUserSession :one
SELECT
   	s.user_id,
	s.refresh_token,
	s.device_name,
	s.expires_at,
   	u.role_id
FROM sessions s
JOIN users u ON s.user_id = u.id
WHERE refresh_token = @refresh_token;
