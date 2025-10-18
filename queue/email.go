package queue

import (
	"context"

	config2 "encore.app/config"
	"encore.app/connection"
	"encore.app/event"
	"encore.dev/pubsub"
)

var _ = pubsub.NewSubscription(
	event.ForgotsEvent, "send-password-reset",
	pubsub.SubscriptionConfig[*event.ForgotEvent]{
		Handler: SendPasswordResetEmail,
	},
)
var _ = pubsub.NewSubscription(
	event.SignupsEvent, "send-otp",
	pubsub.SubscriptionConfig[*event.SignupEvent]{
		Handler: SendSignupEmail,
	},
)

func SendSignupEmail(ctx context.Context, event *event.SignupEvent) error {
	if event == nil {
		return nil
	}
	config := config2.GetConfig()
	redis := connection.GetRedisClient(&config.Redis)
	redis.SetEx(ctx, "otp:"+event.Email, "123456", 5*60)
	return nil
}

func SendPasswordResetEmail(ctx context.Context, event *event.ForgotEvent) error {
	config := config2.GetConfig()
	redis := connection.GetRedisClient(&config.Redis)
	redis.SetEx(ctx, "reset:"+event.Email, "654321", 5*60)
	return nil
}
