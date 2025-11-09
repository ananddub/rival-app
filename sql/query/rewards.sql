-- name: GetAllActiveRewards :many
SELECT * FROM rewards WHERE is_active = true ORDER BY created_at DESC;

-- name: GetRewardById :one
SELECT * FROM rewards WHERE id = $1;

-- name: CreateNewReward :one
INSERT INTO rewards (title, description, type, coins, is_active)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserRewards :many
SELECT * FROM user_rewards WHERE user_id = $1 ORDER BY created_at DESC;

-- name: ClaimUserReward :one
INSERT INTO user_rewards (user_id, reward_id, claimed, claimed_at)
VALUES ($1, $2, true, NOW())
ON CONFLICT (user_id, reward_id) 
DO UPDATE SET claimed = true, claimed_at = NOW()
RETURNING *;

-- name: GetClaimedDailyRewards :many
SELECT * FROM daily_rewards WHERE user_id = $1 ORDER BY day;

-- name: ClaimDailyReward :one
INSERT INTO daily_rewards (user_id, day)
VALUES ($1, $2)
ON CONFLICT (user_id, day) DO NOTHING
RETURNING *;

-- name: GetReferralCount :one
SELECT COUNT(*) FROM referral_rewards WHERE referrer_id = $1;

-- name: AddReferralReward :one
INSERT INTO referral_rewards (referrer_id, referred_user_id)
VALUES ($1, $2)
RETURNING *;
