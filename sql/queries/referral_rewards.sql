-- name: CreateReferralReward :one
INSERT INTO referral_rewards (
    referrer_id, referred_id, reward_amount, reward_type, status
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: ProcessReferralBonus :exec
UPDATE referral_rewards SET 
    status = 'credited',
    credited_at = NOW()
WHERE id = $1;

-- name: GetPendingReferrals :many
SELECT * FROM referral_rewards 
WHERE status = 'pending' 
ORDER BY created_at ASC;

-- name: GetUserReferralStats :one
SELECT 
    COUNT(*) as total_referrals,
    COALESCE(SUM(reward_amount), 0) as total_earned
FROM referral_rewards 
WHERE referrer_id = $1 
AND status = 'credited';
