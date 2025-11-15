-- name: StoreBlacklistedToken :exec
INSERT INTO blacklisted_tokens (user_id, hashed_token)
VALUES ($1, $2);

-- name: IsTokenBlacklisted :one
SELECT EXISTS(SELECT 1 FROM blacklisted_tokens WHERE hashed_token = $1);

-- name: CleanupExpiredTokens :exec
DELETE FROM blacklisted_tokens 
WHERE created_at < NOW() - INTERVAL '24 hours';
