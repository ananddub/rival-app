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
	if err := notificationService.SendNotification(ctx, req.UserID, req.Title, req.Message, req.Type); err != nil {
		return nil, err
	}
	return &SendNotificationResponse{Message: "Notification sent successfully"}, nil
}

type Notification struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	IsRead    bool   `json:"is_read"`
	CreatedAt string `json:"created_at"`
}

type GetNotificationsResponse struct {
	Notifications []Notification `json:"notifications"`
}

//encore:api public method=GET path=/notifications/:userID/:page
func GetNotifications(ctx context.Context, userID int64, page int) (*GetNotificationsResponse, error) {
	notifications, err := notificationService.GetUserNotifications(ctx, userID, page)
	if err != nil {
		return nil, err
	}

	result := make([]Notification, len(notifications))
	for i, n := range notifications {
		result[i] = Notification{
			ID:        n.ID,
			Title:     n.Title,
			Message:   n.Message,
			Type:      n.Type,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return &GetNotificationsResponse{Notifications: result}, nil
}

type MarkAsReadRequest struct {
	NotificationID int64 `json:"notification_id"`
}

type MarkAsReadResponse struct {
	Message string `json:"message"`
}

//encore:api public method=POST path=/notifications/read
func MarkAsRead(ctx context.Context, req *MarkAsReadRequest) (*MarkAsReadResponse, error) {
	if err := notificationService.MarkAsRead(ctx, req.NotificationID); err != nil {
		return nil, err
	}
	return &MarkAsReadResponse{Message: "Notification marked as read"}, nil
}
