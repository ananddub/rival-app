-- name: CreateUser :one
INSERT INTO
    users (
        email,
        password_hash,
        phone,
        name,
        profile_pic,
        firebase_uid,
        coin_balance,
        role,
        referral_code,
        referred_by
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10
    )
RETURNING
    *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByReferralCode :one
SELECT * FROM users WHERE referral_code = $1;

-- name: UpdateUser :exec
UPDATE users
SET
    name = $2,
    phone = $3,
    profile_pic = $4,
    updated_at = NOW()
WHERE
    id = $1;

-- name: UpdateUserPassword :exec
UPDATE users
SET
    password_hash = $2,
    updated_at = NOW()
WHERE
    id = $1;

-- name: CreateJWTSession :exec
INSERT INTO
    jwt_sessions (
        user_id,
        token_hash,
        refresh_token_hash,
        expires_at
    )
VALUES ($1, $2, $3, $4);

-- name: GetJWTSession :one
SELECT *
FROM jwt_sessions
WHERE
    token_hash = $1
    AND is_revoked = false;

-- name: RevokeJWTSession :exec
UPDATE jwt_sessions SET is_revoked = true WHERE token_hash = $1;

-- name: RevokeAllUserSessions :exec
UPDATE jwt_sessions SET is_revoked = true WHERE user_id = $1;
