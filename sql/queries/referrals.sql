-- name: GetReferralRewards :many
SELECT * FROM referral_rewards 
WHERE referrer_id = $1 
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUserPendingReferrals :many
SELECT * FROM referral_rewards 
WHERE referrer_id = $1 AND status = 'pending'
ORDER BY created_at DESC;
