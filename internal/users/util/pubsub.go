package util

import (
	"time"
	"strconv"

	userspb "encore.app/gen/proto/proto/api"
	schemapb "encore.app/gen/proto/proto/schema"
	"encore.app/pkg/pubsub"
)

type UserPubSubService interface {
	PublishWalletUpdate(userID string, newBalance float64, transaction *schemapb.Transaction, eventType string)
	PublishUserNotification(userID, title, message, notificationType string)
	SubscribeWalletUpdates(userID string) *pubsub.Channel
	SubscribeUserNotifications(userID string) *pubsub.Channel
}

type userPubSubService struct {
	ps *pubsub.PubSub
}

func NewUserPubSubService() UserPubSubService {
	return &userPubSubService{
		ps: pubsub.Get(),
	}
}

func (s *userPubSubService) PublishWalletUpdate(userID string, newBalance float64, transaction *schemapb.Transaction, eventType string) {
	topic := "wallet_updates:" + userID
	update := &userspb.StreamWalletUpdatesResponse{
		NewBalance:  newBalance,
		Transaction: transaction,
		EventType:   eventType,
	}
	s.ps.Publish(topic, update)
}

func (s *userPubSubService) PublishUserNotification(userID, title, message, notificationType string) {
	topic := "user_notifications:" + userID
	notification := &userspb.StreamUserNotificationsResponse{
		Id:        generateNotificationID(),
		Title:     title,
		Message:   message,
		Type:      notificationType,
		Timestamp: getCurrentTimestamp(),
	}
	s.ps.Publish(topic, notification)
}

func (s *userPubSubService) SubscribeWalletUpdates(userID string) *pubsub.Channel {
	topic := "wallet_updates:" + userID
	return s.ps.Subscribe(topic)
}

func (s *userPubSubService) SubscribeUserNotifications(userID string) *pubsub.Channel {
	topic := "user_notifications:" + userID
	return s.ps.Subscribe(topic)
}

func generateNotificationID() string {
	// Simple ID generation using timestamp
	return "notif_" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}
