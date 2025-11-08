package reward

import (
	"context"
	"time"
)

type Reward struct {
	ID          int64
	UserID      int64
	Title       string
	Description *string
	Points      int
	Type        string
	Status      string
	ClaimedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Repository interface {
	CreateReward(ctx context.Context, reward *Reward) (int64, error)
	GetReward(ctx context.Context, id int64) (*Reward, error)
	GetUserRewards(ctx context.Context, userID int64, limit, offset int) ([]*Reward, error)
	GetRewardsByType(ctx context.Context, userID int64, rewardType string) ([]*Reward, error)
	UpdateRewardStatus(ctx context.Context, id int64, status string) error
	ClaimReward(ctx context.Context, id int64) error
}

type Service interface {
	CreateReward(ctx context.Context, userID int64, title, description, rewardType string, points int) error
	GetAvailableRewards(ctx context.Context, userID int64) ([]*Reward, error)
	ClaimReward(ctx context.Context, userID, rewardID int64) error
}
