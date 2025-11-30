package util

import (
	"time"
	"strconv"

	paymentpb "rival/gen/proto/proto/api"
	schemapb "rival/gen/proto/proto/schema"
	"rival/pkg/pubsub"
)

type PaymentPubSubService interface {
	PublishPaymentUpdate(userID, paymentID, status, eventType string, coinsAdded float64, purchase *schemapb.CoinPurchase)
	PublishTransactionUpdate(userID, transactionID, status, eventType string, amount float64, transaction *schemapb.Transaction)
	SubscribePaymentUpdates(userID string) *pubsub.Channel
	SubscribeTransactionUpdates(userID string) *pubsub.Channel
}

type paymentPubSubService struct {
	ps *pubsub.PubSub
}

func NewPaymentPubSubService() PaymentPubSubService {
	return &paymentPubSubService{
		ps: pubsub.Get(),
	}
}

func (s *paymentPubSubService) PublishPaymentUpdate(userID, paymentID, status, eventType string, coinsAdded float64, purchase *schemapb.CoinPurchase) {
	topic := "payment_updates:" + userID
	update := &paymentpb.StreamPaymentUpdatesResponse{
		PaymentId:  paymentID,
		Status:     status,
		CoinsAdded: coinsAdded,
		Purchase:   purchase,
		EventType:  eventType,
	}
	s.ps.Publish(topic, update)
}

func (s *paymentPubSubService) PublishTransactionUpdate(userID, transactionID, status, eventType string, amount float64, transaction *schemapb.Transaction) {
	topic := "transaction_updates:" + userID
	update := &paymentpb.StreamTransactionUpdatesResponse{
		TransactionId: transactionID,
		Status:        status,
		Amount:        amount,
		Transaction:   transaction,
		EventType:     eventType,
	}
	s.ps.Publish(topic, update)
}

func (s *paymentPubSubService) SubscribePaymentUpdates(userID string) *pubsub.Channel {
	topic := "payment_updates:" + userID
	return s.ps.Subscribe(topic)
}

func (s *paymentPubSubService) SubscribeTransactionUpdates(userID string) *pubsub.Channel {
	topic := "transaction_updates:" + userID
	return s.ps.Subscribe(topic)
}

func generateNotificationID() string {
	return "notif_" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}
