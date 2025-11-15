-- name: CreateTransaction :one
INSERT INTO transactions (
    user_id, merchant_id, coins_spent, original_amount, 
    transaction_type, status
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: CreateCoinPurchase :one
INSERT INTO coin_purchases (
    user_id, amount, coins_received, payment_method, status
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetUserDailySpending :one
SELECT COALESCE(SUM(coins_spent), 0) as daily_spent
FROM transactions 
WHERE user_id = $1 
AND DATE(created_at) = CURRENT_DATE 
AND status = 'completed';

-- name: GetUserMonthlySpending :one
SELECT COALESCE(SUM(coins_spent), 0) as monthly_spent
FROM transactions 
WHERE user_id = $1 
AND EXTRACT(YEAR FROM created_at) = EXTRACT(YEAR FROM CURRENT_DATE)
AND EXTRACT(MONTH FROM created_at) = EXTRACT(MONTH FROM CURRENT_DATE)
AND status = 'completed';

-- name: GetTransactionByID :one
SELECT * FROM transactions WHERE id = $1;

-- name: UpdateTransactionStatus :exec
UPDATE transactions SET 
    status = $2,
    updated_at = NOW()
WHERE id = $1;
