package rewardService

import (
	"context"
	"fmt"
	"time"

	"encore.app/internal/reward/gen"
	"encore.app/internal/reward/repo"
)

// CreateDailyReward creates a daily login reward for user
func CreateDailyReward(ctx context.Context, userID int64) error {
	title := "Daily Login Bonus"
	description := "Reward for logging in today"
	points := int32(10)
	rewardType := "daily"
	
	_, err := rewardRepo.CreateReward(ctx, userID, title, description, points, rewardType, "available")
	return err
}

// CreateWelcomeReward creates welcome reward for new users
func CreateWelcomeReward(ctx context.Context, userID int64) error {
	title := "Welcome Bonus"
	description := "Welcome to our platform! Here's your bonus"
	points := int32(100)
	rewardType := "welcome"
	
	_, err := rewardRepo.CreateReward(ctx, userID, title, description, points, rewardType, "available")
	return err
}

// CreateReferralReward creates referral reward
func CreateReferralReward(ctx context.Context, userID int64, referredUserName string) error {
	title := "Referral Bonus"
	description := fmt.Sprintf("You referred %s and earned bonus points!", referredUserName)
	points := int32(50)
	rewardType := "referral"
	
	_, err := rewardRepo.CreateReward(ctx, userID, title, description, points, rewardType, "available")
	return err
}

// CreateTransactionReward creates reward for transactions
func CreateTransactionReward(ctx context.Context, userID int64, transactionAmount float64) error {
	title := "Transaction Reward"
	description := fmt.Sprintf("Earned reward for transaction of $%.2f", transactionAmount)
	// 1% of transaction amount as points
	points := int32(transactionAmount)
	rewardType := "transaction"
	
	_, err := rewardRepo.CreateReward(ctx, userID, title, description, points, rewardType, "available")
	return err
}

// ClaimUserReward claims a reward for user
func ClaimUserReward(ctx context.Context, rewardID int64, userID int64) error {
	// First check if reward exists and belongs to user
	reward, err := rewardRepo.GetRewardByID(ctx, rewardID)
	if err != nil {
		return fmt.Errorf("reward not found: %v", err)
	}
	
	if reward.UserID != userID {
		return fmt.Errorf("reward does not belong to user")
	}
	
	if reward.Status != "available" {
		return fmt.Errorf("reward is not available for claiming")
	}
	
	// Claim the reward
	_, err = rewardRepo.ClaimReward(ctx, rewardID)
	return err
}

// GetUserTotalPoints calculates total points earned by user
func GetUserTotalPoints(ctx context.Context, userID int64) (int32, error) {
	rewards, err := rewardRepo.GetRewardsByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}
	
	var totalPoints int32
	for _, reward := range rewards {
		if reward.Status == "claimed" {
			totalPoints += reward.Points
		}
	}
	
	return totalPoints, nil
}

// GetUserAvailableRewards gets all available rewards for user
func GetUserAvailableRewards(ctx context.Context, userID int64) ([]rewardGen.Reward, error) {
	allRewards, err := rewardRepo.GetRewardsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	var availableRewards []rewardGen.Reward
	for _, reward := range allRewards {
		if reward.Status == "available" {
			availableRewards = append(availableRewards, reward)
		}
	}
	
	return availableRewards, nil
}

// CheckDailyRewardEligibility checks if user can get daily reward
func CheckDailyRewardEligibility(ctx context.Context, userID int64) (bool, error) {
	rewards, err := rewardRepo.GetRewardsByType(ctx, "daily")
	if err != nil {
		return false, err
	}
	
	today := time.Now().Format("2006-01-02")
	
	// Check if user already got daily reward today
	for _, reward := range rewards {
		if reward.UserID == userID {
			rewardDate := reward.CreatedAt.Time.Format("2006-01-02")
			if rewardDate == today {
				return false, nil // Already got today's reward
			}
		}
	}
	
	return true, nil
}

// ProcessRewardExpiry marks expired rewards as expired
func ProcessRewardExpiry(ctx context.Context) error {
	// This would be called by a cron job
	// For now, just a placeholder for future implementation
	return nil
}
