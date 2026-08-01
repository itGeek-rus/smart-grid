package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/itGeek-rus/smart-grid.git/internal/repository"

	"github.com/itGeek-rus/smart-grid.git/internal/config"
)

type Consumer struct {
	reader *kafka.Reader
	log    *slog.Logger
}

func NewConsumer(cfg config.KafkaConfig, topic string, log *slog.Logger) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        cfg.Brokers,
			GroupID:        cfg.ConsumeGroup,
			Topic:          topic,
			MinBytes:       1,
			MaxBytes:       10e6,
			CommitInterval: time.Second,
			StartOffset:    kafka.FirstOffset,
		}),
		log: log,
	}
}

func (c *Consumer) Subscribe(
	ctx context.Context,
	_ []string,
	_ string,
	handler repository.MessageHandler,
) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("fetch message: %w", err)
		}

		if err := handler(ctx, msg.Key, msg.Value); err != nil {
			c.log.Error("handler failed",
				slog.String("topic", msg.Topic),
				slog.Int("partition", msg.Partition),
				slog.Int64("offset", msg.Offset),
				slog.String("error", err.Error()),
			)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("commit: %w", err)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
