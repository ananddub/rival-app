-- name: CreateActivity :one
INSERT INTO activities (user_id, action, details, category, icon)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetActivitiesByUserID :many
SELECT * FROM activities WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetActivitiesByCategory :many
SELECT * FROM activities WHERE category = $1 ORDER BY created_at DESC;

-- name: GetActivityByID :one
SELECT * FROM activities WHERE id = $1;

-- name: DeleteActivity :exec
DELETE FROM activities WHERE id = $1;
