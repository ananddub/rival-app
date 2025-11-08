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

type CreateRewardRequest struct {
	UserID      int64  `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Points      int    `json:"points"`
}

type CreateRewardResponse struct {
	Message string `json:"message"`
}

//encore:api public method=POST path=/rewards/create
func CreateReward(ctx context.Context, req *CreateRewardRequest) (*CreateRewardResponse, error) {
	if err := rewardService.CreateReward(ctx, req.UserID, req.Title, req.Description, req.Type, req.Points); err != nil {
		return nil, err
	}
	return &CreateRewardResponse{Message: "Reward created successfully"}, nil
}

type Reward struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Points      int    `json:"points"`
	Type        string `json:"type"`
	Status      string `json:"status"`
}

type GetRewardsResponse struct {
	Rewards []Reward `json:"rewards"`
}

//encore:api public method=GET path=/rewards/:userID
func GetRewards(ctx context.Context, userID int64) (*GetRewardsResponse, error) {
	rewards, err := rewardService.GetAvailableRewards(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]Reward, len(rewards))
	for i, r := range rewards {
		desc := ""
		if r.Description != nil {
			desc = *r.Description
		}
		result[i] = Reward{
			ID:          r.ID,
			Title:       r.Title,
			Description: desc,
			Points:      r.Points,
			Type:        r.Type,
			Status:      r.Status,
		}
	}
	return &GetRewardsResponse{Rewards: result}, nil
}

type ClaimRewardRequest struct {
	UserID   int64 `json:"user_id"`
	RewardID int64 `json:"reward_id"`
}

type ClaimRewardResponse struct {
	Message string `json:"message"`
}

//encore:api public method=POST path=/rewards/claim
func ClaimReward(ctx context.Context, req *ClaimRewardRequest) (*ClaimRewardResponse, error) {
	if err := rewardService.ClaimReward(ctx, req.UserID, req.RewardID); err != nil {
		return nil, err
	}
	return &ClaimRewardResponse{Message: "Reward claimed successfully"}, nil
}
