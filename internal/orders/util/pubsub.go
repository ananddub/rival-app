package util

import (
	"strconv"
	"time"

	orderpb "rival/gen/proto/proto/api"
	schemapb "rival/gen/proto/proto/schema"
	"rival/pkg/pubsub"
)

type OrderPubSubService interface {
	PublishOrderUpdate(userID int, order *schemapb.Order, eventType string)
	SubscribeOrderUpdates(userID int) *pubsub.Channel
}

type orderPubSubService struct {
	ps *pubsub.PubSub
}

func NewOrderPubSubService() OrderPubSubService {
	return &orderPubSubService{
		ps: pubsub.Get(),
	}
}

func (s *orderPubSubService) PublishOrderUpdate(userID int, order *schemapb.Order, eventType string) {
	topic := "order_updates:" + strconv.Itoa(userID)
	update := &orderpb.StreamOrderUpdatesResponse{
		Order:     order,
		EventType: eventType,
	}
	s.ps.Publish(topic, update)
}

func (s *orderPubSubService) SubscribeOrderUpdates(userID int) *pubsub.Channel {
	topic := "order_updates:" + strconv.Itoa(userID)
	return s.ps.Subscribe(topic)
}

func generateNotificationID() string {
	return "notif_" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}
