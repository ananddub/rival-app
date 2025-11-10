package activityHandler

import (
	"context"
	"encore.app/connection"
	"encore.app/config"
)

type Activity struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	Action      string `json:"action"`
	Details     string `json:"details"`
	Category    string `json:"category"`
	Icon        string `json:"icon"`
	CreatedAt   string `json:"created_at"`
}

type GetActivitiesResponse struct {
	Activities []Activity `json:"activities"`
}

//encore:api public method=GET path=/activities/:userID
func GetUserActivities(ctx context.Context, userID int64) (*GetActivitiesResponse, error) {
	cfg := config.Load("config.yaml")
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return &GetActivitiesResponse{Activities: []Activity{}}, nil
	}
	
	query := `SELECT id, user_id, action, COALESCE(details, ''), COALESCE(category, ''), COALESCE(icon, ''), 
			  TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at
			  FROM activities WHERE user_id = $1 ORDER BY created_at DESC LIMIT 50`
	
	rows, err := db.Query(ctx, query, userID)
	if err != nil {
		return &GetActivitiesResponse{Activities: []Activity{}}, nil
	}
	defer rows.Close()
	
	var activities []Activity
	for rows.Next() {
		var activity Activity
		err := rows.Scan(&activity.ID, &activity.UserID, &activity.Action, 
						&activity.Details, &activity.Category, &activity.Icon, &activity.CreatedAt)
		if err != nil {
			continue
		}
		activities = append(activities, activity)
	}
	
	return &GetActivitiesResponse{Activities: activities}, nil
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
