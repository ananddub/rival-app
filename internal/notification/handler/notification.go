package notificationHandler

import (
	"context"
	"strconv"

	"encore.app/internal/notification/repo"
)

type CreateNotificationRequest struct {
	UserID  string `json:"user_id"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

type NotificationResponse struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Type    string `json:"type"`
	IsRead  bool   `json:"is_read"`
}

type GetNotificationsResponse struct {
	Notifications []*NotificationResponse `json:"notifications"`
}

//encore:api public method=POST path=/notification
func CreateNotification(ctx context.Context, req *CreateNotificationRequest) (*NotificationResponse, error) {
	userID, _ := strconv.ParseInt(req.UserID, 10, 64)
	
	notification, err := notificationRepo.CreateNotification(ctx, userID, req.Title, req.Message, req.Type)
	if err != nil {
		return nil, err
	}

	return &NotificationResponse{
		ID:      strconv.FormatInt(notification.ID, 10),
		UserID:  strconv.FormatInt(notification.UserID, 10),
		Title:   notification.Title,
		Message: notification.Message,
		Type:    notification.Type,
		IsRead:  notification.IsRead,
	}, nil
}

//encore:api public method=GET path=/notification/user/:userID
func GetUserNotifications(ctx context.Context, userID string) (*GetNotificationsResponse, error) {
	uid, _ := strconv.ParseInt(userID, 10, 64)
	
	notifications, err := notificationRepo.GetNotificationsByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}

	var result []*NotificationResponse
	for _, notification := range notifications {
		result = append(result, &NotificationResponse{
			ID:      strconv.FormatInt(notification.ID, 10),
			UserID:  strconv.FormatInt(notification.UserID, 10),
			Title:   notification.Title,
			Message: notification.Message,
			Type:    notification.Type,
			IsRead:  notification.IsRead,
		})
	}

	return &GetNotificationsResponse{Notifications: result}, nil
}

//encore:api public method=PUT path=/notification/:id/read
func MarkAsRead(ctx context.Context, id string) (*NotificationResponse, error) {
	notificationID, _ := strconv.ParseInt(id, 10, 64)
	
	notification, err := notificationRepo.MarkAsRead(ctx, notificationID)
	if err != nil {
		return nil, err
	}

	return &NotificationResponse{
		ID:      strconv.FormatInt(notification.ID, 10),
		UserID:  strconv.FormatInt(notification.UserID, 10),
		Title:   notification.Title,
		Message: notification.Message,
		Type:    notification.Type,
		IsRead:  notification.IsRead,
	}, nil
}
