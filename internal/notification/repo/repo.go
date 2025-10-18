package notificationRepo

import (
	"context"

	"encore.app/config"
	"encore.app/connection"
	notification_gen "encore.app/internal/notification/gen"
	"github.com/jackc/pgx/v5/pgtype"
)

type Notification struct {
	ID      int64
	UserID  int64
	Title   string
	Message string
	Type    string
	IsRead  bool
}

func CreateNotification(ctx context.Context, userID int64, title, message, notificationType string) (*Notification, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := notification_gen.New(db)
	
	notification, err := queries.CreateNotification(ctx, notification_gen.CreateNotificationParams{
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    pgtype.Text{String: notificationType, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return &Notification{
		ID:      notification.ID,
		UserID:  notification.UserID,
		Title:   notification.Title,
		Message: notification.Message,
		Type:    notification.Type.String,
		IsRead:  notification.IsRead.Bool,
	}, nil
}

func GetNotificationsByUserID(ctx context.Context, userID int64) ([]*Notification, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := notification_gen.New(db)
	
	notifications, err := queries.GetNotificationsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []*Notification
	for _, notification := range notifications {
		result = append(result, &Notification{
			ID:      notification.ID,
			UserID:  notification.UserID,
			Title:   notification.Title,
			Message: notification.Message,
			Type:    notification.Type.String,
			IsRead:  notification.IsRead.Bool,
		})
	}

	return result, nil
}

func MarkAsRead(ctx context.Context, notificationID int64) (*Notification, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := notification_gen.New(db)
	
	notification, err := queries.MarkAsRead(ctx, notificationID)
	if err != nil {
		return nil, err
	}

	return &Notification{
		ID:      notification.ID,
		UserID:  notification.UserID,
		Title:   notification.Title,
		Message: notification.Message,
		Type:    notification.Type.String,
		IsRead:  notification.IsRead.Bool,
	}, nil
}
