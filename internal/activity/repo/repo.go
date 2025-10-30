package activityRepo

import (
	"context"

	"encore.app/config"
	"encore.app/connection"
	activity_gen "encore.app/internal/activity/gen"
)

type Activity struct {
	ID        int64
	UserID    int64
	Action    string
	Details   string
	Category  string
	Icon      string
	CreatedAt string
}

func CreateActivity(ctx context.Context, userID int64, action, details, category, icon string) (*Activity, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := activity_gen.New(db)

	activity, err := queries.CreateActivity(ctx, activity_gen.CreateActivityParams{
		UserID:   userID,
		Action:   action,
		Details:  details,
		Category: category,
		Icon:     icon,
	})
	if err != nil {
		return nil, err
	}

	return &Activity{
		ID:        activity.ID,
		UserID:    activity.UserID,
		Action:    activity.Action,
		Details:   activity.Details,
		Category:  activity.Category,
		Icon:      activity.Icon,
		CreatedAt: activity.CreatedAt.Time.String(),
	}, nil
}

func GetActivitiesByUserID(ctx context.Context, userID int64) ([]*Activity, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := activity_gen.New(db)

	activities, err := queries.GetActivitiesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []*Activity
	for _, activity := range activities {
		result = append(result, &Activity{
			ID:        activity.ID,
			UserID:    activity.UserID,
			Action:    activity.Action,
			Details:   activity.Details,
			Category:  activity.Category,
			Icon:      activity.Icon,
			CreatedAt: activity.CreatedAt.Time.String(),
		})
	}

	return result, nil
}
