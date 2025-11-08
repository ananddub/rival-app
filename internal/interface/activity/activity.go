package activity

import (
	"context"
	"time"
)

type Activity struct {
	ID        int64
	UserID    int64
	Action    string
	Details   string
	Category  string
	Icon      string
	CreatedAt time.Time
}

type Repository interface {
	CreateActivity(ctx context.Context, activity *Activity) error
	GetActivities(ctx context.Context, userID int64, limit, offset int) ([]*Activity, error)
	GetActivitiesByCategory(ctx context.Context, userID int64, category string, limit, offset int) ([]*Activity, error)
}

type Service interface {
	LogActivity(ctx context.Context, userID int64, action, details, category, icon string) error
	GetUserActivities(ctx context.Context, userID int64, page int) ([]*Activity, error)
}
