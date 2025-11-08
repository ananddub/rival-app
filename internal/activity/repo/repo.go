package repo

import (
	"context"

	"encore.app/config"
	"encore.app/connection"
	db "encore.app/gen"
	activityInterface "encore.app/internal/interface/activity"
)

type ActivityRepo struct{}

func New() activityInterface.Repository {
	return &ActivityRepo{}
}

func (r *ActivityRepo) CreateActivity(ctx context.Context, activity *activityInterface.Activity) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	_, err = queries.CreateActivity(ctx, db.CreateActivityParams{
		UserID:   activity.UserID,
		Action:   activity.Action,
		Details:  activity.Details,
		Category: activity.Category,
		Icon:     activity.Icon,
	})
	return err
}

func (r *ActivityRepo) GetActivities(ctx context.Context, userID int64, limit, offset int) ([]*activityInterface.Activity, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	activities, err := queries.GetActivitiesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	end := offset + limit
	if end > len(activities) {
		end = len(activities)
	}
	if offset >= len(activities) {
		return []*activityInterface.Activity{}, nil
	}

	result := make([]*activityInterface.Activity, 0)
	for i := offset; i < end; i++ {
		a := activities[i]
		result = append(result, &activityInterface.Activity{
			ID:        a.ID,
			UserID:    a.UserID,
			Action:    a.Action,
			Details:   a.Details,
			Category:  a.Category,
			Icon:      a.Icon,
			CreatedAt: a.CreatedAt.Time,
		})
	}
	return result, nil
}

func (r *ActivityRepo) GetActivitiesByCategory(ctx context.Context, userID int64, category string, limit, offset int) ([]*activityInterface.Activity, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	activities, err := queries.GetActivitiesByCategory(ctx, category)
	if err != nil {
		return nil, err
	}

	result := make([]*activityInterface.Activity, 0)
	for _, a := range activities {
		if a.UserID != userID {
			continue
		}
		result = append(result, &activityInterface.Activity{
			ID:        a.ID,
			UserID:    a.UserID,
			Action:    a.Action,
			Details:   a.Details,
			Category:  a.Category,
			Icon:      a.Icon,
			CreatedAt: a.CreatedAt.Time,
		})
	}

	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	if offset >= len(result) {
		return []*activityInterface.Activity{}, nil
	}

	return result[offset:end], nil
}
