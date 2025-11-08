package activityHandler

import (
	"context"

	"encore.app/internal/activity/repo"
	"encore.app/internal/activity/service"
)

var (
	activityRepo    = repo.New()
	activityService = service.New(activityRepo)
)

type LogActivityRequest struct {
	UserID   int64  `json:"user_id"`
	Action   string `json:"action"`
	Details  string `json:"details"`
	Category string `json:"category"`
	Icon     string `json:"icon"`
}

type LogActivityResponse struct {
	Message string `json:"message"`
}

//encore:api public method=POST path=/activities/log
func LogActivity(ctx context.Context, req *LogActivityRequest) (*LogActivityResponse, error) {
	if err := activityService.LogActivity(ctx, req.UserID, req.Action, req.Details, req.Category, req.Icon); err != nil {
		return nil, err
	}
	return &LogActivityResponse{Message: "Activity logged successfully"}, nil
}

type Activity struct {
	ID        int64  `json:"id"`
	Action    string `json:"action"`
	Details   string `json:"details"`
	Category  string `json:"category"`
	Icon      string `json:"icon"`
	CreatedAt string `json:"created_at"`
}

type GetActivitiesResponse struct {
	Activities []Activity `json:"activities"`
}

//encore:api public method=GET path=/activities/:userID/:page
func GetActivities(ctx context.Context, userID int64, page int) (*GetActivitiesResponse, error) {
	activities, err := activityService.GetUserActivities(ctx, userID, page)
	if err != nil {
		return nil, err
	}

	result := make([]Activity, len(activities))
	for i, a := range activities {
		result[i] = Activity{
			ID:        a.ID,
			Action:    a.Action,
			Details:   a.Details,
			Category:  a.Category,
			Icon:      a.Icon,
			CreatedAt: a.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return &GetActivitiesResponse{Activities: result}, nil
}
