-- name: CreateJWTToken :one
INSERT INTO jwt_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetJWTToken :one
SELECT * FROM jwt_tokens WHERE token_hash = $1 AND expires_at > now();

-- name: DeleteJWTToken :exec
DELETE FROM jwt_tokens WHERE token_hash = $1;

-- name: DeleteExpiredTokens :exec
DELETE FROM jwt_tokens WHERE expires_at <= now();
