package service

import (
	"context"
	"errors"

	rewardInterface "encore.app/internal/interface/reward"
)

type RewardService struct {
	repo rewardInterface.Repository
}

func New(repo rewardInterface.Repository) rewardInterface.Service {
	return &RewardService{repo: repo}
}

func (s *RewardService) CreateReward(ctx context.Context, userID int64, title, description, rewardType string, points int) error {
	desc := description
	reward := &rewardInterface.Reward{
		UserID:      userID,
		Title:       title,
		Description: &desc,
		Points:      points,
		Type:        rewardType,
		Status:      "available",
	}

	_, err := s.repo.CreateReward(ctx, reward)
	return err
}

func (s *RewardService) GetAvailableRewards(ctx context.Context, userID int64) ([]*rewardInterface.Reward, error) {
	return s.repo.GetUserRewards(ctx, userID, 100, 0)
}

func (s *RewardService) ClaimReward(ctx context.Context, userID, rewardID int64) error {
	reward, err := s.repo.GetReward(ctx, rewardID)
	if err != nil {
		return err
	}

	if reward.UserID != userID {
		return errors.New("unauthorized")
	}

	if reward.Status != "available" {
		return errors.New("reward not available")
	}

	if err := s.repo.ClaimReward(ctx, rewardID); err != nil {
		return err
	}

	return s.repo.UpdateRewardStatus(ctx, rewardID, "claimed")
}
