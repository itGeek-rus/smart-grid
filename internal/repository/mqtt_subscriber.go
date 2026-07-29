package repository

import "context"

type MQTTMessage struct {
	Topic   string
	Payload []byte
}

type MQTTHandler func(ctx context.Context, msg MQTTMessage) error

type MQTTSubscriber interface {
	Subscribe(ctx context.Context, topic string, qos byte, handler MQTTHandler) error
	Close() error
}
