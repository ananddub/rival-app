package activityHandler

import "context"

type Activity struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	Action      string `json:"action"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

type GetActivitiesResponse struct {
	Activities []Activity `json:"activities"`
}

//encore:api public method=GET path=/activities/:userID
func GetUserActivities(ctx context.Context, userID int64) (*GetActivitiesResponse, error) {
	return &GetActivitiesResponse{Activities: []Activity{}}, nil
}

type LogActivityRequest struct {
	UserID      int64  `json:"user_id"`
	Action      string `json:"action"`
	Description string `json:"description"`
}

type LogActivityResponse struct {
	Message string `json:"message"`
}

//encore:api public method=POST path=/activities/log
func LogActivity(ctx context.Context, req *LogActivityRequest) (*LogActivityResponse, error) {
	return &LogActivityResponse{Message: "Activity logged successfully"}, nil
}
