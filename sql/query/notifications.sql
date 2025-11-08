-- name: CreateNotification :one
INSERT INTO notifications (user_id, title, message, type)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetNotificationsByUserID :many
SELECT * FROM notifications WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetUnreadNotifications :many
SELECT * FROM notifications WHERE user_id = $1 AND is_read = FALSE ORDER BY created_at DESC;

-- name: MarkAsRead :one
UPDATE notifications SET is_read = TRUE WHERE id = $1 RETURNING *;

-- name: DeleteNotification :exec
DELETE FROM notifications WHERE id = $1;
