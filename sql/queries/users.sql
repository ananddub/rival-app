-- name: GetUserProfile :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateUserProfile :exec
UPDATE users
SET
    name = $2,
    phone = $3,
    profile_pic = $4,
    updated_at = NOW()
WHERE
    id = $1;

-- name: UpdateCoinBalance :exec
UPDATE users
SET
    coin_balance = $2,
    updated_at = NOW()
WHERE
    id = $1;

-- name: GetUserTransactions :many
SELECT *
FROM transactions
WHERE
    user_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3;

-- name: GetUserCoinPurchases :many
SELECT *
FROM coin_purchases
WHERE
    user_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3;

-- name: GetUserReferralRewards :many
SELECT *
FROM referral_rewards
WHERE
    referrer_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3;

-- name: UpdateReferralRewardStatus :exec
UPDATE referral_rewards
SET
    status = $2,
    credited_at = CASE
        WHEN $2 = 'credited' THEN NOW()
        ELSE credited_at
    END
WHERE
    id = $1;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: GetAllUsers :many
SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: DleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: DeleteByEmail :exec
DELETE FROM users WHERE email = $1;
