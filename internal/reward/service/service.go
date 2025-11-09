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

func (s *RewardService) GetAllRewards(ctx context.Context) ([]*rewardInterface.Reward, error) {
	return s.repo.GetRewards(ctx)
}

func (s *RewardService) GetUserRewards(ctx context.Context, userID int64) ([]*rewardInterface.UserReward, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}
	return s.repo.GetUserRewards(ctx, userID)
}

func (s *RewardService) ClaimReward(ctx context.Context, userID, rewardID int64) (*rewardInterface.UserReward, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}
	if rewardID <= 0 {
		return nil, errors.New("invalid reward ID")
	}

	// Check if reward exists and is active
	reward, err := s.repo.GetRewardByID(ctx, rewardID)
	if err != nil {
		return nil, err
	}
	if !reward.IsActive {
		return nil, errors.New("reward is not active")
	}

	// Check if already claimed
	userRewards, err := s.repo.GetUserRewards(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, ur := range userRewards {
		if ur.RewardID == rewardID && ur.Claimed {
			return nil, errors.New("reward already claimed")
		}
	}

	return s.repo.ClaimReward(ctx, userID, rewardID)
}

func (s *RewardService) GetDailyRewards(ctx context.Context, userID int64) ([]*rewardInterface.DailyReward, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}
	return s.repo.GetDailyRewards(ctx, userID)
}

func (s *RewardService) ClaimDailyReward(ctx context.Context, userID int64, day int) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}
	if day < 1 || day > 7 {
		return errors.New("invalid day. Must be between 1-7")
	}

	// Check if previous days are claimed (sequential claiming)
	dailyRewards, err := s.repo.GetDailyRewards(ctx, userID)
	if err != nil {
		return err
	}

	// Check if current day is already claimed
	for _, dr := range dailyRewards {
		if dr.Day == day && dr.Claimed {
			return errors.New("daily reward already claimed for this day")
		}
	}

	// Check sequential claiming (must claim previous days first)
	if day > 1 {
		prevDayClaimed := false
		for _, dr := range dailyRewards {
			if dr.Day == day-1 && dr.Claimed {
				prevDayClaimed = true
				break
			}
		}
		if !prevDayClaimed {
			return errors.New("must claim previous day's reward first")
		}
	}

	return s.repo.ClaimDailyReward(ctx, userID, day)
}

func (s *RewardService) ProcessReferralReward(ctx context.Context, referrerID, newUserID int64) error {
	if referrerID <= 0 {
		return errors.New("invalid referrer ID")
	}
	if newUserID <= 0 {
		return errors.New("invalid new user ID")
	}
	if referrerID == newUserID {
		return errors.New("cannot refer yourself")
	}

	return s.repo.AddReferralReward(ctx, referrerID)
}

func (s *RewardService) GetReferralStats(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, errors.New("invalid user ID")
	}
	return s.repo.GetReferralRewards(ctx, userID)
}
