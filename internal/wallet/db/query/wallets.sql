-- name: CreateWallet :one
INSERT INTO wallets (user_id, balance, coins, currency)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetWalletByUserID :one
SELECT * FROM wallets WHERE user_id = $1;

-- name: UpdateBalance :one
UPDATE wallets SET balance = $2, updated_at = NOW() WHERE user_id = $1 RETURNING *;

-- name: UpdateCoins :one
UPDATE wallets SET coins = $2, updated_at = NOW() WHERE user_id = $1 RETURNING *;

-- name: CreateTransaction :one
INSERT INTO transactions (user_id, wallet_id, title, description, amount, type, icon)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetTransactionsByUserID :many
SELECT * FROM transactions WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetTransactionsByType :many
SELECT * FROM transactions WHERE user_id = $1 AND type = $2 ORDER BY created_at DESC;
