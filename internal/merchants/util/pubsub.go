package util

import (
	"time"
	"strconv"

	merchantpb "rival/gen/proto/proto/api"
	schemapb "rival/gen/proto/proto/schema"
	"rival/pkg/pubsub"
)

type MerchantPubSubService interface {
	PublishOrderUpdate(merchantID int, order *schemapb.Order, eventType string)
	PublishMerchantNotification(merchantID int, title, message, notificationType string)
	SubscribeOrderUpdates(merchantID int) *pubsub.Channel
	SubscribeMerchantNotifications(merchantID int) *pubsub.Channel
}

type merchantPubSubService struct {
	ps *pubsub.PubSub
}

func NewMerchantPubSubService() MerchantPubSubService {
	return &merchantPubSubService{
		ps: pubsub.Get(),
	}
}

func (s *merchantPubSubService) PublishOrderUpdate(merchantID int, order *schemapb.Order, eventType string) {
	topic := "merchant_orders:" + strconv.Itoa(merchantID)
	update := &merchantpb.StreamOrdersResponse{
		Order:     order,
		EventType: eventType,
	}
	s.ps.Publish(topic, update)
}

func (s *merchantPubSubService) PublishMerchantNotification(merchantID int, title, message, notificationType string) {
	topic := "merchant_notifications:" + strconv.Itoa(merchantID)
	notification := &merchantpb.StreamNotificationsResponse{
		Id:        generateNotificationID(),
		Title:     title,
		Message:   message,
		Type:      notificationType,
		Timestamp: getCurrentTimestamp(),
	}
	s.ps.Publish(topic, notification)
}

func (s *merchantPubSubService) SubscribeOrderUpdates(merchantID int) *pubsub.Channel {
	topic := "merchant_orders:" + strconv.Itoa(merchantID)
	return s.ps.Subscribe(topic)
}

func (s *merchantPubSubService) SubscribeMerchantNotifications(merchantID int) *pubsub.Channel {
	topic := "merchant_notifications:" + strconv.Itoa(merchantID)
	return s.ps.Subscribe(topic)
}

func generateNotificationID() string {
	return "notif_" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}
