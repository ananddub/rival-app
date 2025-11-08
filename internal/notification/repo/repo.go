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
	})
	return err
}

func (r *NotificationRepo) GetNotifications(ctx context.Context, userID int64, limit, offset int) ([]*notificationInterface.Notification, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	notifications, err := queries.GetNotificationsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	end := offset + limit
	if end > len(notifications) {
		end = len(notifications)
	}
	if offset >= len(notifications) {
		return []*notificationInterface.Notification{}, nil
	}

	result := make([]*notificationInterface.Notification, 0)
	for i := offset; i < end; i++ {
		n := notifications[i]
		result = append(result, &notificationInterface.Notification{
			ID:        n.ID,
			UserID:    n.UserID,
			Title:     n.Title,
			Message:   n.Message,
			Type:      n.Type.String,
			IsRead:    n.IsRead.Bool,
			CreatedAt: n.CreatedAt.Time,
		})
	}
	return result, nil
}

func (r *NotificationRepo) GetUnreadCount(ctx context.Context, userID int64) (int, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return 0, err
	}

	queries := db.New(conn)
	notifications, err := queries.GetUnreadNotifications(ctx, userID)
	return len(notifications), err
}

func (r *NotificationRepo) MarkAsRead(ctx context.Context, id int64) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	_, err = queries.MarkAsRead(ctx, id)
	return err
}

func (r *NotificationRepo) MarkAllAsRead(ctx context.Context, userID int64) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	notifications, err := queries.GetUnreadNotifications(ctx, userID)
	if err != nil {
		return err
	}

	for _, n := range notifications {
		if _, err := queries.MarkAsRead(ctx, n.ID); err != nil {
			return err
		}
	}
	return nil
}
