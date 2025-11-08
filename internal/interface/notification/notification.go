package notification

import (
	"context"
	"time"
)

type Notification struct {
	ID        int64
	UserID    int64
	Title     string
	Message   string
	Type      string
	IsRead    bool
	CreatedAt time.Time
}

type Repository interface {
	CreateNotification(ctx context.Context, notification *Notification) error
	GetNotifications(ctx context.Context, userID int64, limit, offset int) ([]*Notification, error)
	GetUnreadCount(ctx context.Context, userID int64) (int, error)
	MarkAsRead(ctx context.Context, id int64) error
	MarkAllAsRead(ctx context.Context, userID int64) error
}

type Service interface {
	SendNotification(ctx context.Context, userID int64, title, message, notifType string) error
	GetUserNotifications(ctx context.Context, userID int64, page int) ([]*Notification, error)
	MarkAsRead(ctx context.Context, notificationID int64) error
}
