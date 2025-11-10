package reward

import (
	"context"
	"time"
)

type Reward struct {
	ID          int64
	Title       string
	Description string
	Type        string // daily, weekly, achievement, referral
	Coins       int64
	Money       float64
	IsActive    bool
	CreatedAt   time.Time
}

type UserReward struct {
	ID        int64
	UserID    int64
	RewardID  int64
	Claimed   bool
	ClaimedAt *time.Time
	CreatedAt time.Time
}

type DailyReward struct {
	Day     int
	Coins   int64
	Money   float64
	Claimed bool
}

type Repository interface {
	GetRewards(ctx context.Context) ([]*Reward, error)
	GetRewardByID(ctx context.Context, id int64) (*Reward, error)
	CreateReward(ctx context.Context, reward *Reward) (*Reward, error)
	
	GetUserRewards(ctx context.Context, userID int64) ([]*UserReward, error)
	ClaimReward(ctx context.Context, userID, rewardID int64) (*UserReward, error)
	GetDailyRewards(ctx context.Context, userID int64) ([]*DailyReward, error)
	ClaimDailyReward(ctx context.Context, userID int64, day int) error
	
	GetReferralRewards(ctx context.Context, userID int64) (int64, error)
	AddReferralReward(ctx context.Context, userID int64) error
}

type Service interface {
	GetAllRewards(ctx context.Context) ([]*Reward, error)
	GetUserRewards(ctx context.Context, userID int64) ([]*UserReward, error)
	ClaimReward(ctx context.Context, userID, rewardID int64) (*UserReward, error)
	
	GetDailyRewards(ctx context.Context, userID int64) ([]*DailyReward, error)
	ClaimDailyReward(ctx context.Context, userID int64, day int) error
	
	ProcessReferralReward(ctx context.Context, referrerID, newUserID int64) error
	GetReferralStats(ctx context.Context, userID int64) (int64, error)
}
