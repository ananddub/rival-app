-- name: GetUserCoins :one
SELECT coins FROM wallets WHERE user_id = $1;

-- name: AddCoins :exec
UPDATE wallets SET coins = coins + $2, updated_at = NOW() WHERE user_id = $1;

-- name: DeductCoins :exec
UPDATE wallets SET coins = coins - $2, updated_at = NOW() WHERE user_id = $1 AND coins >= $2;

-- name: CreateCoinTransaction :one
INSERT INTO coin_transactions (user_id, coins, type, reason) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetAllCoinTransactions :many
SELECT * FROM coin_transactions WHERE user_id = $1 ORDER BY created_at DESC;

-- name: CreateCoinPackage :one
INSERT INTO coin_packages (name, coins, price, bonus_coins, is_active)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetCoinTransactions :many
SELECT * FROM coin_transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: GetCoinPackages :many
SELECT * FROM coin_packages WHERE is_active = true ORDER BY price ASC;

-- name: CreateCoinPurchase :one
INSERT INTO coin_purchases (user_id, package_id, coins_received, amount_paid, payment_status, payment_id) 
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: UpdatePurchaseStatus :exec
UPDATE coin_purchases SET payment_status = $2 WHERE id = $1;
