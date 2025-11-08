package service

import (
	"context"

	activityInterface "encore.app/internal/interface/activity"
)

type ActivityService struct {
	repo activityInterface.Repository
}

func New(repo activityInterface.Repository) activityInterface.Service {
	return &ActivityService{repo: repo}
}

func (s *ActivityService) LogActivity(ctx context.Context, userID int64, action, details, category, icon string) error {
	activity := &activityInterface.Activity{
		UserID:   userID,
		Action:   action,
		Details:  details,
		Category: category,
		Icon:     icon,
	}
	return s.repo.CreateActivity(ctx, activity)
}

func (s *ActivityService) GetUserActivities(ctx context.Context, userID int64, page int) ([]*activityInterface.Activity, error) {
	limit := 20
	offset := (page - 1) * limit
	return s.repo.GetActivities(ctx, userID, limit, offset)
}
