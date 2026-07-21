package repository

import "context"

type MessageHandler func(ctx context.Context, key, value []byte) error

type EventConsumer interface {
	Subscribe(ctx context.Context, topics []string, group string, handler MessageHandler) error
	Close() error
}
