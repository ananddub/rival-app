package rewardHandler

import (
	"context"
	
	rewardInterface "encore.app/internal/interface/reward"
)

type TestRewardResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

//encore:api public method=GET path=/rewards/test
func TestReward(ctx context.Context) (*TestRewardResponse, error) {
	return &TestRewardResponse{
		Message: "Reward system is working properly",
		Status:  "success",
	}, nil
}

type RewardDashboardResponse struct {
	TotalRewards      int                   `json:"total_rewards"`
	ClaimedRewards    int                   `json:"claimed_rewards"`
	PendingRewards    int                   `json:"pending_rewards"`
	DailyRewards      []DailyRewardResponse `json:"daily_rewards"`
	ReferralStats     ReferralStatsResponse `json:"referral_stats"`
	AvailableRewards  []RewardResponse      `json:"available_rewards"`
}

//encore:api public method=GET path=/rewards/dashboard/:userID
func GetRewardDashboard(ctx context.Context, userID int64) (*RewardDashboardResponse, error) {
	// Get all rewards
	allRewards, err := rewardService.GetAllRewards(ctx)
	if err != nil {
		return nil, err
	}

	// Get user rewards
	userRewards, err := rewardService.GetUserRewards(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get daily rewards
	dailyRewards, err := rewardService.GetDailyRewards(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get referral stats
	referralCount, err := rewardService.GetReferralStats(ctx, userID)
	if err != nil {
		referralCount = 0
	}

	// Count claimed rewards
	claimedCount := 0
	for _, ur := range userRewards {
		if ur.Claimed {
			claimedCount++
		}
	}

	// Convert daily rewards
	dailyRewardResponses := make([]DailyRewardResponse, len(dailyRewards))
	for i, dr := range dailyRewards {
		dailyRewardResponses[i] = DailyRewardResponse{
			Day:     dr.Day,
			Coins:   dr.Coins,
			Money:   dr.Money,
			Claimed: dr.Claimed,
		}
	}

	// Convert available rewards
	availableRewards := make([]RewardResponse, len(allRewards))
	for i, reward := range allRewards {
		availableRewards[i] = RewardResponse{
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

	return &RewardDashboardResponse{
		TotalRewards:   len(allRewards),
		ClaimedRewards: claimedCount,
		PendingRewards: len(allRewards) - claimedCount,
		DailyRewards:   dailyRewardResponses,
		ReferralStats: ReferralStatsResponse{
			TotalReferrals: referralCount,
			RewardEarned:   referralCount * 100,
		},
		AvailableRewards: availableRewards,
	}, nil
}

type CreateRewardRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
	Coins       int64   `json:"coins"`
	Money       float64 `json:"money"`
}

type CreateRewardResponse struct {
	Message string         `json:"message"`
	Reward  RewardResponse `json:"reward"`
}

//encore:api public method=POST path=/rewards/create
func CreateReward(ctx context.Context, req *CreateRewardRequest) (*CreateRewardResponse, error) {
	// This would typically be admin-only
	reward := &rewardInterface.Reward{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Coins:       req.Coins,
		Money:       req.Money,
		IsActive:    true,
	}

	createdReward, err := rewardRepo.CreateReward(ctx, reward)
	if err != nil {
		return nil, err
	}

	return &CreateRewardResponse{
		Message: "Reward created successfully",
		Reward: RewardResponse{
			ID:          createdReward.ID,
			Title:       createdReward.Title,
			Description: createdReward.Description,
			Type:        createdReward.Type,
			Coins:       createdReward.Coins,
			Money:       createdReward.Money,
			IsActive:    createdReward.IsActive,
			CreatedAt:   createdReward.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}, nil
}
