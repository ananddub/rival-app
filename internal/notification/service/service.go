package service

import (
	"context"
	"errors"

	notificationInterface "encore.app/internal/interface/notification"
)

type NotificationService struct {
	repo notificationInterface.Repository
}

func New(repo notificationInterface.Repository) notificationInterface.Service {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) GetUserNotifications(ctx context.Context, userID int64) ([]*notificationInterface.Notification, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}
	return s.repo.GetUserNotifications(ctx, userID)
}

func (s *NotificationService) SendNotification(ctx context.Context, userID int64, title, message, notificationType string) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}
	if title == "" {
		return errors.New("title cannot be empty")
	}
	if message == "" {
		return errors.New("message cannot be empty")
	}

	notification := &notificationInterface.Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    notificationType,
		IsRead:  false,
	}

	return s.repo.CreateNotification(ctx, notification)
}

func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID int64) error {
	if notificationID <= 0 {
		return errors.New("invalid notification ID")
	}
	return s.repo.MarkAsRead(ctx, notificationID)
}
