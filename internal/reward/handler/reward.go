package rewardHandler

import (
	"context"
	"fmt"
	"strconv"

	"encore.app/internal/reward/repo"
	"encore.app/internal/reward/service"
)

type CreateRewardRequest struct {
	UserID      string `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Points      int32  `json:"points"`
	Type        string `json:"type"`
}

type ClaimRewardRequest struct {
	UserID string `json:"user_id"`
}

type RewardResponse struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Points      int32  `json:"points"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	ClaimedAt   string `json:"claimed_at,omitempty"`
}

type GetRewardsResponse struct {
	Rewards []*RewardResponse `json:"rewards"`
}

type UserPointsResponse struct {
	UserID      string `json:"user_id"`
	TotalPoints int32  `json:"total_points"`
}

//encore:api public method=POST path=/reward
func CreateReward(ctx context.Context, req *CreateRewardRequest) (*RewardResponse, error) {
	userID, _ := strconv.ParseInt(req.UserID, 10, 64)

	reward, err := rewardRepo.CreateReward(ctx, userID, req.Title, req.Description, req.Points, req.Type, "available")
	if err != nil {
		return nil, err
	}

	return &RewardResponse{
		ID:          strconv.FormatInt(reward.ID, 10),
		UserID:      strconv.FormatInt(reward.UserID, 10),
		Title:       reward.Title,
		Description: *reward.Description,
		Points:      reward.Points,
		Type:        reward.Type,
		Status:      reward.Status,
	}, nil
}

//encore:api public method=GET path=/reward/user/:userID
func GetUserRewards(ctx context.Context, userID string) (*GetRewardsResponse, error) {
	uid, _ := strconv.ParseInt(userID, 10, 64)

	rewards, err := rewardRepo.GetRewardsByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}

	var result []*RewardResponse
	for _, reward := range rewards {
		claimedAt := ""
		if reward.ClaimedAt != nil {
			claimedAt = reward.ClaimedAt.Format("2006-01-02T15:04:05Z")
		}

		result = append(result, &RewardResponse{
			ID:          strconv.FormatInt(reward.ID, 10),
			UserID:      strconv.FormatInt(reward.UserID, 10),
			Title:       reward.Title,
			Description: *reward.Description,
			Points:      reward.Points,
			Type:        reward.Type,
			Status:      reward.Status,
			ClaimedAt:   claimedAt,
		})
	}

	return &GetRewardsResponse{Rewards: result}, nil
}

//encore:api public method=GET path=/reward/available
func GetAvailableRewards(ctx context.Context) (*GetRewardsResponse, error) {
	rewards, err := rewardRepo.GetAvailableRewards(ctx)
	if err != nil {
		return nil, err
	}

	var result []*RewardResponse
	for _, reward := range rewards {
		result = append(result, &RewardResponse{
			ID:          strconv.FormatInt(reward.ID, 10),
			UserID:      strconv.FormatInt(reward.UserID, 10),
			Title:       reward.Title,
			Description: *reward.Description,
			Points:      reward.Points,
			Type:        reward.Type,
			Status:      reward.Status,
		})
	}

	return &GetRewardsResponse{Rewards: result}, nil
}

//encore:api public method=POST path=/reward/:id/claim
func ClaimReward(ctx context.Context, id string, req *ClaimRewardRequest) (*RewardResponse, error) {
	rewardID, _ := strconv.ParseInt(id, 10, 64)
	userID, _ := strconv.ParseInt(req.UserID, 10, 64)

	// Use service layer for business logic
	err := rewardService.ClaimUserReward(ctx, rewardID, userID)
	if err != nil {
		return nil, err
	}

	// Get updated reward
	reward, err := rewardRepo.GetRewardByID(ctx, rewardID)
	if err != nil {
		return nil, err
	}

	claimedAt := ""
	if reward.ClaimedAt != nil {
		claimedAt = reward.ClaimedAt.Format("2006-01-02T15:04:05Z")
	}

	return &RewardResponse{
		ID:          strconv.FormatInt(reward.ID, 10),
		UserID:      strconv.FormatInt(reward.UserID, 10),
		Title:       reward.Title,
		Description: *reward.Description,
		Points:      reward.Points,
		Type:        reward.Type,
		Status:      reward.Status,
		ClaimedAt:   claimedAt,
	}, nil
}

//encore:api public method=POST path=/reward/:userID/daily
func CreateDailyReward(ctx context.Context, userID string) (*RewardResponse, error) {
	uid, _ := strconv.ParseInt(userID, 10, 64)

	// Check eligibility first
	eligible, err := rewardService.CheckDailyRewardEligibility(ctx, uid)
	if err != nil {
		return nil, err
	}

	if !eligible {
		return nil, fmt.Errorf("daily reward already claimed today")
	}

	// Create daily reward using service
	err = rewardService.CreateDailyReward(ctx, uid)
	if err != nil {
		return nil, err
	}

	// Get the created reward
	rewards, err := rewardRepo.GetRewardsByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}

	// Return the latest daily reward
	for i := len(rewards) - 1; i >= 0; i-- {
		if rewards[i].Type == "daily" {
			reward := rewards[i]
			return &RewardResponse{
				ID:          strconv.FormatInt(reward.ID, 10),
				UserID:      strconv.FormatInt(reward.UserID, 10),
				Title:       reward.Title,
				Description: *reward.Description,
				Points:      reward.Points,
				Type:        reward.Type,
				Status:      reward.Status,
			}, nil
		}
	}

	return nil, fmt.Errorf("failed to create daily reward")
}

//encore:api public method=GET path=/reward/points/:userID
func GetUserPoints(ctx context.Context, userID string) (*UserPointsResponse, error) {
	uid, _ := strconv.ParseInt(userID, 10, 64)

	totalPoints, err := rewardService.GetUserTotalPoints(ctx, uid)
	if err != nil {
		return nil, err
	}

	return &UserPointsResponse{
		UserID:      userID,
		TotalPoints: totalPoints,
	}, nil
}
