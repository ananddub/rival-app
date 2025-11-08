-- name: CreateUser :one
INSERT INTO users (full_name, email, phone_number, password_hash, sign_type, role)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateUser :one
UPDATE users 
SET full_name = $2, email = $3, phone_number = $4, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdatePassword :one
UPDATE users
SET password_hash = $2, updated_at = now()
WHERE id = $1
RETURNING *;


-- name: UpdateEmailVerification :one
UPDATE users SET is_email_verified = $2, updated_at = now() WHERE id = $1 RETURNING *;

-- name: UpdatePhoneVerification :one
UPDATE users SET is_phone_verified = $2, updated_at = now() WHERE id = $1 RETURNING *;
