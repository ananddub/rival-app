package service

import (
	"context"

	notificationInterface "encore.app/internal/interface/notification"
)

type NotificationService struct {
	repo notificationInterface.Repository
}

func New(repo notificationInterface.Repository) notificationInterface.Service {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) SendNotification(ctx context.Context, userID int64, title, message, notifType string) error {
	notification := &notificationInterface.Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    notifType,
		IsRead:  false,
	}
	return s.repo.CreateNotification(ctx, notification)
}

func (s *NotificationService) GetUserNotifications(ctx context.Context, userID int64, page int) ([]*notificationInterface.Notification, error) {
	limit := 20
	offset := (page - 1) * limit
	return s.repo.GetNotifications(ctx, userID, limit, offset)
}

func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID int64) error {
	return s.repo.MarkAsRead(ctx, notificationID)
}
