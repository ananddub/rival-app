package rewardHandler

import (
	"context"

	"encore.app/internal/reward/repo"
	"encore.app/internal/reward/service"
)

var (
	rewardRepo    = repo.New()
	rewardService = service.New(rewardRepo)
)

type RewardResponse struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
	Coins       int64   `json:"coins"`
	Money       float64 `json:"money"`
	IsActive    bool    `json:"is_active"`
	CreatedAt   string  `json:"created_at"`
}

type GetRewardsResponse struct {
	Rewards []RewardResponse `json:"rewards"`
}

//encore:api public method=GET path=/rewards
func GetRewards(ctx context.Context) (*GetRewardsResponse, error) {
	rewards, err := rewardService.GetAllRewards(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]RewardResponse, len(rewards))
	for i, reward := range rewards {
		result[i] = RewardResponse{
			ID:          reward.ID,
			Title:       reward.Title,
			Description: reward.Description,
			Type:        reward.Type,
			Coins:       reward.Coins,
			Money:       reward.Money,
			IsActive:    reward.IsActive,
			CreatedAt:   reward.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return &GetRewardsResponse{Rewards: result}, nil
}

type UserRewardResponse struct {
	ID        int64   `json:"id"`
	RewardID  int64   `json:"reward_id"`
	Title     string  `json:"title"`
	Coins     int64   `json:"coins"`
	Money     float64 `json:"money"`
	Claimed   bool    `json:"claimed"`
	ClaimedAt string  `json:"claimed_at,omitempty"`
}

type GetUserRewardsResponse struct {
	UserRewards []UserRewardResponse `json:"user_rewards"`
}

//encore:api public method=GET path=/rewards/user/:userID
func GetUserRewards(ctx context.Context, userID int64) (*GetUserRewardsResponse, error) {
	userRewards, err := rewardService.GetUserRewards(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]UserRewardResponse, len(userRewards))
	for i, ur := range userRewards {
		claimedAt := ""
		if ur.ClaimedAt != nil {
			claimedAt = ur.ClaimedAt.Format("2006-01-02 15:04:05")
		}

		result[i] = UserRewardResponse{
			ID:        ur.ID,
			RewardID:  ur.RewardID,
			Claimed:   ur.Claimed,
			ClaimedAt: claimedAt,
		}
	}

	return &GetUserRewardsResponse{UserRewards: result}, nil
}

type ClaimRewardRequest struct {
	UserID   int64 `json:"user_id"`
	RewardID int64 `json:"reward_id"`
}

type ClaimRewardResponse struct {
	Message     string  `json:"message"`
	CoinsEarned int64   `json:"coins_earned"`
	MoneyEarned float64 `json:"money_earned"`
}

//encore:api public method=POST path=/rewards/claim
func ClaimReward(ctx context.Context, req *ClaimRewardRequest) (*ClaimRewardResponse, error) {
	_, err := rewardService.ClaimReward(ctx, req.UserID, req.RewardID)
	if err != nil {
		return nil, err
	}

	return &ClaimRewardResponse{
		Message:     "Reward claimed successfully",
		CoinsEarned: 0,
		MoneyEarned: 0,
	}, nil
}

type DailyRewardResponse struct {
	Day     int     `json:"day"`
	Coins   int64   `json:"coins"`
	Money   float64 `json:"money"`
	Claimed bool    `json:"claimed"`
}

type GetDailyRewardsResponse struct {
	DailyRewards []DailyRewardResponse `json:"daily_rewards"`
	CurrentDay   int                   `json:"current_day"`
}

//encore:api public method=GET path=/rewards/daily/:userID
func GetDailyRewards(ctx context.Context, userID int64) (*GetDailyRewardsResponse, error) {
	dailyRewards, err := rewardService.GetDailyRewards(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]DailyRewardResponse, len(dailyRewards))
	currentDay := 1

	for i, dr := range dailyRewards {
		result[i] = DailyRewardResponse{
			Day:     dr.Day,
			Coins:   dr.Coins,
			Money:   dr.Money,
			Claimed: dr.Claimed,
		}
		if dr.Claimed {
			currentDay = dr.Day + 1
		}
	}

	return &GetDailyRewardsResponse{
		DailyRewards: result,
		CurrentDay:   currentDay,
	}, nil
}

type ClaimDailyRewardRequest struct {
	UserID int64 `json:"user_id"`
	Day    int   `json:"day"`
}

type ClaimDailyRewardResponse struct {
	Message     string  `json:"message"`
	CoinsEarned int64   `json:"coins_earned"`
	MoneyEarned float64 `json:"money_earned"`
	NextDay     int     `json:"next_day"`
}

//encore:api public method=POST path=/rewards/daily/claim
func ClaimDailyReward(ctx context.Context, req *ClaimDailyRewardRequest) (*ClaimDailyRewardResponse, error) {
	if err := rewardService.ClaimDailyReward(ctx, req.UserID, req.Day); err != nil {
		return nil, err
	}

	// Get daily reward details
	dailyRewards, err := rewardService.GetDailyRewards(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	var coins int64
	var money float64
	for _, dr := range dailyRewards {
		if dr.Day == req.Day {
			coins = dr.Coins
			money = dr.Money
			break
		}
	}

	return &ClaimDailyRewardResponse{
		Message:     "Daily reward claimed successfully",
		CoinsEarned: coins,
		MoneyEarned: money,
		NextDay:     req.Day + 1,
	}, nil
}

type ReferralStatsResponse struct {
	TotalReferrals int64 `json:"total_referrals"`
	RewardEarned   int64 `json:"reward_earned"`
}

//encore:api public method=GET path=/rewards/referral/:userID
func GetReferralStats(ctx context.Context, userID int64) (*ReferralStatsResponse, error) {
	totalReferrals, err := rewardService.GetReferralStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &ReferralStatsResponse{
		TotalReferrals: totalReferrals,
		RewardEarned:   totalReferrals * 100, // 100 coins per referral
	}, nil
}

type ProcessReferralRequest struct {
	ReferrerID int64 `json:"referrer_id"`
	NewUserID  int64 `json:"new_user_id"`
}

type ProcessReferralResponse struct {
	Message string `json:"message"`
}

//encore:api public method=POST path=/rewards/referral
func ProcessReferral(ctx context.Context, req *ProcessReferralRequest) (*ProcessReferralResponse, error) {
	if err := rewardService.ProcessReferralReward(ctx, req.ReferrerID, req.NewUserID); err != nil {
		return nil, err
	}

	return &ProcessReferralResponse{
		Message: "Referral reward processed successfully",
	}, nil
}
