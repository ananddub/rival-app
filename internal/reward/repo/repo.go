package repo

import (
	"context"
	"time"

	"encore.app/config"
	"encore.app/connection"
	db "encore.app/gen"
	rewardInterface "encore.app/internal/interface/reward"
	"github.com/jackc/pgx/v5/pgtype"
)

type RewardRepo struct{}

func New() rewardInterface.Repository {
	return &RewardRepo{}
}

func (r *RewardRepo) GetRewards(ctx context.Context) ([]*rewardInterface.Reward, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	rewards, err := queries.GetAllActiveRewards(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*rewardInterface.Reward, len(rewards))
	for i, reward := range rewards {
		result[i] = &rewardInterface.Reward{
			ID:          reward.ID,
			Title:       reward.Title,
			Description: reward.Description.String,
			Type:        reward.Type,
			Coins:       reward.Coins.Int64,
			Money:       0, // Will be added if needed
			IsActive:    reward.IsActive.Bool,
			CreatedAt:   reward.CreatedAt.Time,
		}
	}
	return result, nil
}

func (r *RewardRepo) GetRewardByID(ctx context.Context, id int64) (*rewardInterface.Reward, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	reward, err := queries.GetRewardById(ctx, id)
	if err != nil {
		return nil, err
	}

	return &rewardInterface.Reward{
		ID:          reward.ID,
		Title:       reward.Title,
		Description: reward.Description.String,
		Type:        reward.Type,
		Coins:       reward.Coins.Int64,
		Money:       0,
		IsActive:    reward.IsActive.Bool,
		CreatedAt:   reward.CreatedAt.Time,
	}, nil
}

func (r *RewardRepo) CreateReward(ctx context.Context, reward *rewardInterface.Reward) (*rewardInterface.Reward, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	newReward, err := queries.CreateNewReward(ctx, db.CreateNewRewardParams{
		Title:       reward.Title,
		Description: pgtype.Text{String: reward.Description, Valid: true},
		Type:        reward.Type,
		Coins:       pgtype.Int8{Int64: reward.Coins, Valid: true},
		IsActive:    pgtype.Bool{Bool: reward.IsActive, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return &rewardInterface.Reward{
		ID:          newReward.ID,
		Title:       newReward.Title,
		Description: newReward.Description.String,
		Type:        newReward.Type,
		Coins:       newReward.Coins.Int64,
		Money:       0,
		IsActive:    newReward.IsActive.Bool,
		CreatedAt:   newReward.CreatedAt.Time,
	}, nil
}

func (r *RewardRepo) GetUserRewards(ctx context.Context, userID int64) ([]*rewardInterface.UserReward, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	userRewards, err := queries.GetUserRewards(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*rewardInterface.UserReward, len(userRewards))
	for i, ur := range userRewards {
		var claimedAt *time.Time
		if ur.ClaimedAt.Valid {
			claimedAt = &ur.ClaimedAt.Time
		}

		result[i] = &rewardInterface.UserReward{
			ID:        ur.ID,
			UserID:    ur.UserID,
			RewardID:  ur.RewardID,
			Claimed:   ur.Claimed.Bool,
			ClaimedAt: claimedAt,
			CreatedAt: ur.CreatedAt.Time,
		}
	}
	return result, nil
}

func (r *RewardRepo) ClaimReward(ctx context.Context, userID, rewardID int64) (*rewardInterface.UserReward, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	userReward, err := queries.ClaimUserReward(ctx, db.ClaimUserRewardParams{
		UserID:   userID,
		RewardID: rewardID,
	})
	if err != nil {
		return nil, err
	}

	var claimedAt *time.Time
	if userReward.ClaimedAt.Valid {
		claimedAt = &userReward.ClaimedAt.Time
	}

	return &rewardInterface.UserReward{
		ID:        userReward.ID,
		UserID:    userReward.UserID,
		RewardID:  userReward.RewardID,
		Claimed:   userReward.Claimed.Bool,
		ClaimedAt: claimedAt,
		CreatedAt: userReward.CreatedAt.Time,
	}, nil
}

func (r *RewardRepo) GetDailyRewards(ctx context.Context, userID int64) ([]*rewardInterface.DailyReward, error) {
	// Hardcoded daily rewards for 7 days
	dailyRewards := []*rewardInterface.DailyReward{
		{Day: 1, Coins: 10, Money: 0, Claimed: false},
		{Day: 2, Coins: 20, Money: 0, Claimed: false},
		{Day: 3, Coins: 30, Money: 0, Claimed: false},
		{Day: 4, Coins: 40, Money: 0, Claimed: false},
		{Day: 5, Coins: 50, Money: 0, Claimed: false},
		{Day: 6, Coins: 75, Money: 0, Claimed: false},
		{Day: 7, Coins: 100, Money: 5.0, Claimed: false},
	}

	// Check which days are claimed (this would be stored in database)
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return dailyRewards, nil // Return default if DB error
	}

	queries := db.New(conn)
	claimedDays, err := queries.GetClaimedDailyRewards(ctx, userID)
	if err != nil {
		return dailyRewards, nil // Return default if error
	}

	// Mark claimed days
	for _, claimedDay := range claimedDays {
		day := int(claimedDay.Day)
		if day >= 1 && day <= 7 {
			dailyRewards[day-1].Claimed = true
		}
	}

	return dailyRewards, nil
}

func (r *RewardRepo) ClaimDailyReward(ctx context.Context, userID int64, day int) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	_, err = queries.ClaimDailyReward(ctx, db.ClaimDailyRewardParams{
		UserID: userID,
		Day:    int32(day),
	})
	return err
}

func (r *RewardRepo) GetReferralRewards(ctx context.Context, userID int64) (int64, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return 0, err
	}

	queries := db.New(conn)
	count, err := queries.GetReferralCount(ctx, userID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *RewardRepo) AddReferralReward(ctx context.Context, userID int64) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	_, err = queries.AddReferralReward(ctx, db.AddReferralRewardParams{
		ReferrerID:     userID,
		ReferredUserID: 0, // This should be passed as parameter
	})
	return err
}
