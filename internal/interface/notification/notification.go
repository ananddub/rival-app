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
	GetUserNotifications(ctx context.Context, userID int64) ([]*Notification, error)
	CreateNotification(ctx context.Context, notification *Notification) error
	MarkAsRead(ctx context.Context, notificationID int64) error
}

type Service interface {
	GetUserNotifications(ctx context.Context, userID int64) ([]*Notification, error)
	SendNotification(ctx context.Context, userID int64, title, message, notificationType string) error
	MarkAsRead(ctx context.Context, notificationID int64) error
}
