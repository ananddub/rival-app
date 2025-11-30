-- name: CreateOrder :one
INSERT INTO orders (
    merchant_id, user_id, offer_id, order_number, items, subtotal, discount_amount, total_amount, coins_used, status, notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetOrderByID :one
SELECT * FROM orders WHERE id = $1;

-- name: GetOrderByNumber :one
SELECT * FROM orders WHERE order_number = $1;

-- name: UpdateOrderStatus :exec
UPDATE orders SET
    status = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: GetUserOrders :many
SELECT * FROM orders 
WHERE user_id = $1 
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;

-- name: GetUserOrdersByStatus :many
SELECT * FROM orders 
WHERE user_id = $1 AND status = $2
ORDER BY created_at DESC 
LIMIT $3 OFFSET $4;

-- name: GetMerchantOrders :many
SELECT * FROM orders 
WHERE merchant_id = $1 
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;

-- name: GetMerchantOrdersByStatus :many
SELECT * FROM orders 
WHERE merchant_id = $1 AND status = $2
ORDER BY created_at DESC 
LIMIT $3 OFFSET $4;

-- name: CountUserOrders :one
SELECT COUNT(*) FROM orders WHERE user_id = $1;

-- name: CountMerchantOrders :one
SELECT COUNT(*) FROM orders WHERE merchant_id = $1;
