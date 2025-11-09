-- name: GetUserNotifications :many
SELECT * FROM notifications WHERE user_id = $1 ORDER BY created_at DESC;

-- name: CreateNotification :one
INSERT INTO notifications (user_id, title, message, type, is_read)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: MarkNotificationAsRead :exec
UPDATE notifications SET is_read = true WHERE id = $1;
