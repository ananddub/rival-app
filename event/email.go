package event

import "encore.dev/pubsub"

type SignupEvent struct{ Email string }
type ForgotEvent struct {
	Email string
}

var SignupsEvent = pubsub.NewTopic[*SignupEvent]("signups", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

var ForgotsEvent = pubsub.NewTopic[*ForgotEvent]("forgot", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
