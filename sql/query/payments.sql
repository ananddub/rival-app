-- name: CreatePayment :one
INSERT INTO payments (user_id, order_id, amount, currency, status, payment_method, payment_gateway, payment_type, reference_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetPaymentByOrderID :one
SELECT * FROM payments WHERE order_id = $1;

-- name: UpdatePaymentStatus :one
UPDATE payments SET status = $2, payment_id = $3, updated_at = NOW() WHERE order_id = $1 RETURNING *;

-- name: GetUserPayments :many
SELECT * FROM payments WHERE user_id = $1 ORDER BY created_at DESC;
