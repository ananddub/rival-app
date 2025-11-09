package repo

import (
	"context"

	"encore.app/config"
	"encore.app/connection"
	db "encore.app/gen"
	notificationInterface "encore.app/internal/interface/notification"
	"github.com/jackc/pgx/v5/pgtype"
)

type NotificationRepo struct{}

func New() notificationInterface.Repository {
	return &NotificationRepo{}
}

func (r *NotificationRepo) GetUserNotifications(ctx context.Context, userID int64) ([]*notificationInterface.Notification, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	notifications, err := queries.GetUserNotifications(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*notificationInterface.Notification, len(notifications))
	for i, n := range notifications {
		result[i] = &notificationInterface.Notification{
			ID:        n.ID,
			UserID:    n.UserID,
			Title:     n.Title,
			Message:   n.Message,
			Type:      n.Type.String,
			IsRead:    n.IsRead.Bool,
			CreatedAt: n.CreatedAt.Time,
		}
	}
	return result, nil
}

func (r *NotificationRepo) CreateNotification(ctx context.Context, notification *notificationInterface.Notification) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	_, err = queries.CreateNotification(ctx, db.CreateNotificationParams{
		UserID:  notification.UserID,
		Title:   notification.Title,
		Message: notification.Message,
		Type:    pgtype.Text{String: notification.Type, Valid: true},
		IsRead:  pgtype.Bool{Bool: notification.IsRead, Valid: true},
	})
	return err
}

func (r *NotificationRepo) MarkAsRead(ctx context.Context, notificationID int64) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	return queries.MarkNotificationAsRead(ctx, notificationID)
}
