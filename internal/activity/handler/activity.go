package activityHandler

import (
	"context"
	"strconv"

	"encore.app/internal/activity/repo"
)

type CreateActivityRequest struct {
	UserID   string `json:"user_id"`
	Action   string `json:"action"`
	Details  string `json:"details"`
	Category string `json:"category"`
	Icon     string `json:"icon"`
}

type ActivityResponse struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Action   string `json:"action"`
	Details  string `json:"details"`
	Category string `json:"category"`
	Icon     string `json:"icon"`
}

type GetActivitiesResponse struct {
	Activities []*ActivityResponse `json:"activities"`
}

//encore:api public method=POST path=/activity
func CreateActivity(ctx context.Context, req *CreateActivityRequest) (*ActivityResponse, error) {
	userID, _ := strconv.ParseInt(req.UserID, 10, 64)
	
	activity, err := activityRepo.CreateActivity(ctx, userID, req.Action, req.Details, req.Category, req.Icon)
	if err != nil {
		return nil, err
	}

	return &ActivityResponse{
		ID:       strconv.FormatInt(activity.ID, 10),
		UserID:   strconv.FormatInt(activity.UserID, 10),
		Action:   activity.Action,
		Details:  activity.Details,
		Category: activity.Category,
		Icon:     activity.Icon,
	}, nil
}

//encore:api public method=GET path=/activity/user/:userID
func GetUserActivities(ctx context.Context, userID string) (*GetActivitiesResponse, error) {
	uid, _ := strconv.ParseInt(userID, 10, 64)
	
	activities, err := activityRepo.GetActivitiesByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}

	var result []*ActivityResponse
	for _, activity := range activities {
		result = append(result, &ActivityResponse{
			ID:       strconv.FormatInt(activity.ID, 10),
			UserID:   strconv.FormatInt(activity.UserID, 10),
			Action:   activity.Action,
			Details:  activity.Details,
			Category: activity.Category,
			Icon:     activity.Icon,
		})
	}

	return &GetActivitiesResponse{Activities: result}, nil
}
