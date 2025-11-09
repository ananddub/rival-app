package notificationHandler

import (
	"context"

	"encore.app/internal/notification/repo"
	"encore.app/internal/notification/service"
)

var (
	notificationRepo    = repo.New()
	notificationService = service.New(notificationRepo)
)

type NotificationResponse struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	IsRead    bool   `json:"is_read"`
	CreatedAt string `json:"created_at"`
}

type GetNotificationsResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
}

//encore:api public method=GET path=/notifications/:userID
func GetNotifications(ctx context.Context, userID int64) (*GetNotificationsResponse, error) {
	notifications, err := notificationService.GetUserNotifications(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		result[i] = NotificationResponse{
			ID:        n.ID,
			UserID:    n.UserID,
			Title:     n.Title,
			Message:   n.Message,
			Type:      n.Type,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return &GetNotificationsResponse{Notifications: result}, nil
}

type SendNotificationRequest struct {
	UserID  int64  `json:"user_id"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

type SendNotificationResponse struct {
	Message string `json:"message"`
}

//encore:api public method=POST path=/notifications/send
func SendNotification(ctx context.Context, req *SendNotificationRequest) (*SendNotificationResponse, error) {
	err := notificationService.SendNotification(ctx, req.UserID, req.Title, req.Message, req.Type)
	if err != nil {
		return nil, err
	}

	return &SendNotificationResponse{Message: "Notification sent successfully"}, nil
}

type MarkAsReadResponse struct {
	Message string `json:"message"`
}

//encore:api public method=PUT path=/notifications/read/:notificationID
func MarkAsRead(ctx context.Context, notificationID int64) (*MarkAsReadResponse, error) {
	err := notificationService.MarkAsRead(ctx, notificationID)
	if err != nil {
		return nil, err
	}

	return &MarkAsReadResponse{Message: "Notification marked as read"}, nil
}
