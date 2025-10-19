package rewardRepo

import (
	"context"

	"encore.app/config"
	"encore.app/connection"
	"encore.app/internal/reward/gen"
)

func CreateReward(ctx context.Context, userID int64, title, description string, points int32, rewardType, status string) (rewardGen.Reward, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return rewardGen.Reward{}, err
	}
	queries := rewardGen.New(db)
	return queries.CreateReward(ctx, rewardGen.CreateRewardParams{
		UserID:      userID,
		Title:       title,
		Description: &description,
		Points:      points,
		Type:        rewardType,
		Status:      status,
	})
}

func GetRewardByID(ctx context.Context, id int64) (rewardGen.Reward, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return rewardGen.Reward{}, err
	}
	queries := rewardGen.New(db)
	return queries.GetRewardByID(ctx, id)
}

func GetRewardsByUserID(ctx context.Context, userID int64) ([]rewardGen.Reward, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}
	queries := rewardGen.New(db)
	return queries.GetRewardsByUserID(ctx, userID)
}

func GetRewardsByType(ctx context.Context, rewardType string) ([]rewardGen.Reward, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}
	queries := rewardGen.New(db)
	return queries.GetRewardsByType(ctx, rewardType)
}

func GetAvailableRewards(ctx context.Context) ([]rewardGen.Reward, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}
	queries := rewardGen.New(db)
	return queries.GetAvailableRewards(ctx)
}

func ClaimReward(ctx context.Context, id int64) (rewardGen.Reward, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return rewardGen.Reward{}, err
	}
	queries := rewardGen.New(db)
	return queries.ClaimReward(ctx, id)
}

func GetClaimedRewards(ctx context.Context, userID int64) ([]rewardGen.Reward, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}
	queries := rewardGen.New(db)
	return queries.GetClaimedRewards(ctx, userID)
}
