package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/itGeek-rus/smart-grid.git/internal/pkg/metrics"

	"github.com/itGeek-rus/smart-grid.git/internal/config"
	"github.com/itGeek-rus/smart-grid.git/internal/domain"
)

type Producer struct {
	rawWriter       *kafka.Writer
	processedWriter *kafka.Writer
	alertsWriter    *kafka.Writer
	commandWriter   *kafka.Writer
	dlqWriter       *kafka.Writer
}

func NewProducer(cfg config.KafkaConfig) *Producer {
	newWriter := func(topic string) *kafka.Writer {
		return &kafka.Writer{
			Addr:         kafka.TCP(cfg.Brokers...),
			Topic:        topic,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireOne,
			Async:        false,
			BatchTimeout: 10 * time.Millisecond,
		}
	}

	return &Producer{
		rawWriter:       newWriter(cfg.TopicRawTelemetry),
		processedWriter: newWriter(cfg.TopicProcessedEvents),
		alertsWriter:    newWriter(cfg.TopicAlerts),
		commandWriter:   newWriter(cfg.TopicCommands),
		dlqWriter:       newWriter(cfg.TopicDLQ),
	}
}

func (p *Producer) PublishRawTelemetry(ctx context.Context, event domain.RawTelemetryEvent) error {
	return p.publish(ctx, p.rawWriter, event.DeviceID, event)
}

func (p *Producer) PublishProcessed(ctx context.Context, event domain.ProcessedTelemetryEvent) error {
	return p.publish(ctx, p.processedWriter, event.DeviceID, event)
}

func (p *Producer) PublishAlert(ctx context.Context, event domain.AlertEvent) error {
	return p.publish(ctx, p.alertsWriter, event.DeviceID, event)
}

func (p *Producer) PublishCommand(ctx context.Context, event domain.DeviceCommandEvent) error {
	return p.publish(ctx, p.commandWriter, event.DeviceID, event)
}

func (p *Producer) PublishDLQ(ctx context.Context, event domain.DLQEvent) error {
	key := event.EventID
	if key == "" {
		key = "dlq"
	}
	return p.publish(ctx, p.dlqWriter, key, event)
}

func (p *Producer) publish(ctx context.Context, w *kafka.Writer, key string, v any) error {
	body, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	err = w.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: body,
		Time:  time.Now().UTC(),
	})
	if err != nil {
		metrics.KafkaPublished.WithLabelValues("raw.telemetry", "error").Inc()
		return fmt.Errorf("kafka write topic=%s: %w", w.Topic, err)
	}
	metrics.KafkaPublished.WithLabelValues("raw.telemetry", "ok").Inc()
	return nil
}

func (p *Producer) Close() error {
	var first error
	for _, w := range []*kafka.Writer{p.rawWriter, p.processedWriter, p.alertsWriter, p.commandWriter, p.dlqWriter} {
		if w == nil {
			continue
		}
		if err := w.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}
