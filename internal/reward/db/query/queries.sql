-- name: CreateReward :one
INSERT INTO rewards (user_id, title, description, points, type, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetRewardByID :one
SELECT * FROM rewards WHERE id = $1;

-- name: GetRewardsByUserID :many
SELECT * FROM rewards WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetRewardsByType :many
SELECT * FROM rewards WHERE type = $1 ORDER BY created_at DESC;

-- name: GetAvailableRewards :many
SELECT * FROM rewards WHERE status = 'available' ORDER BY created_at DESC;

-- name: ClaimReward :one
UPDATE rewards SET status = 'claimed', claimed_at = CURRENT_TIMESTAMP
WHERE id = $1 AND status = 'available'
RETURNING *;

-- name: GetClaimedRewards :many
SELECT * FROM rewards WHERE user_id = $1 AND status = 'claimed' ORDER BY claimed_at DESC;
