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

func (r *RewardRepo) CreateReward(ctx context.Context, reward *rewardInterface.Reward) (int64, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return 0, err
	}

	queries := db.New(conn)
	result, err := queries.CreateReward(ctx, db.CreateRewardParams{
		UserID:      reward.UserID,
		Title:       reward.Title,
		Description: pgtype.Text{String: *reward.Description, Valid: reward.Description != nil},
		Points:      int32(reward.Points),
		Type:        reward.Type,
		Status:      reward.Status,
	})
	if err != nil {
		return 0, err
	}
	return result.ID, nil
}

func (r *RewardRepo) GetReward(ctx context.Context, id int64) (*rewardInterface.Reward, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	reward, err := queries.GetRewardByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var desc *string
	if reward.Description.Valid {
		desc = &reward.Description.String
	}
	var claimedAt *time.Time
	if reward.ClaimedAt.Valid {
		claimedAt = &reward.ClaimedAt.Time
	}

	return &rewardInterface.Reward{
		ID:          reward.ID,
		UserID:      reward.UserID,
		Title:       reward.Title,
		Description: desc,
		Points:      int(reward.Points),
		Type:        reward.Type,
		Status:      reward.Status,
		ClaimedAt:   claimedAt,
		CreatedAt:   reward.CreatedAt.Time,
		UpdatedAt:   reward.UpdatedAt.Time,
	}, nil
}

func (r *RewardRepo) GetUserRewards(ctx context.Context, userID int64, limit, offset int) ([]*rewardInterface.Reward, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	rewards, err := queries.GetRewardsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	end := offset + limit
	if end > len(rewards) {
		end = len(rewards)
	}
	if offset >= len(rewards) {
		return []*rewardInterface.Reward{}, nil
	}

	result := make([]*rewardInterface.Reward, 0)
	for i := offset; i < end; i++ {
		reward := rewards[i]
		var desc *string
		if reward.Description.Valid {
			desc = &reward.Description.String
		}
		var claimedAt *time.Time
		if reward.ClaimedAt.Valid {
			claimedAt = &reward.ClaimedAt.Time
		}
		result = append(result, &rewardInterface.Reward{
			ID:          reward.ID,
			UserID:      reward.UserID,
			Title:       reward.Title,
			Description: desc,
			Points:      int(reward.Points),
			Type:        reward.Type,
			Status:      reward.Status,
			ClaimedAt:   claimedAt,
			CreatedAt:   reward.CreatedAt.Time,
			UpdatedAt:   reward.UpdatedAt.Time,
		})
	}
	return result, nil
}

func (r *RewardRepo) GetRewardsByType(ctx context.Context, userID int64, rewardType string) ([]*rewardInterface.Reward, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	rewards, err := queries.GetRewardsByType(ctx, rewardType)
	if err != nil {
		return nil, err
	}

	result := make([]*rewardInterface.Reward, 0)
	for _, reward := range rewards {
		if reward.UserID != userID {
			continue
		}
		var desc *string
		if reward.Description.Valid {
			desc = &reward.Description.String
		}
		var claimedAt *time.Time
		if reward.ClaimedAt.Valid {
			claimedAt = &reward.ClaimedAt.Time
		}
		result = append(result, &rewardInterface.Reward{
			ID:          reward.ID,
			UserID:      reward.UserID,
			Title:       reward.Title,
			Description: desc,
			Points:      int(reward.Points),
			Type:        reward.Type,
			Status:      reward.Status,
			ClaimedAt:   claimedAt,
			CreatedAt:   reward.CreatedAt.Time,
			UpdatedAt:   reward.UpdatedAt.Time,
		})
	}
	return result, nil
}

func (r *RewardRepo) UpdateRewardStatus(ctx context.Context, id int64, status string) error {
	// Status is updated by ClaimReward query
	return nil
}

func (r *RewardRepo) ClaimReward(ctx context.Context, id int64) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	_, err = queries.ClaimReward(ctx, id)
	return err
}
