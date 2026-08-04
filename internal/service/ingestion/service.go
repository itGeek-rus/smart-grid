package ingestion

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/itGeek-rus/smart-grid.git/internal/pkg/metrics"

	"github.com/itGeek-rus/smart-grid.git/internal/domain"
	"github.com/itGeek-rus/smart-grid.git/internal/repository"
)

type Service struct {
	publisher   repository.EventPublisher
	sourceTopic string
	log         *slog.Logger
}

func NewService(publisher repository.EventPublisher, sourceTopic string, log *slog.Logger) *Service {
	return &Service{
		publisher:   publisher,
		sourceTopic: sourceTopic,
		log:         log,
	}
}

func (s *Service) HandleMQTT(ctx context.Context, msg repository.MQTTMessage) error {
	receivedAt := time.Now().UTC()

	event, err := ParseAndValidate(msg.Topic, msg.Payload, receivedAt)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return s.toDLQ(ctx, msg, err.Error())
		}
		return err
	}

	if err := s.publisher.PublishRawTelemetry(ctx, event); err != nil {
		return err
	}

	metrics.MQTTMessages.WithLabelValues("ok").Inc()

	s.log.Info("telemetry ingested",
		slog.String("device_id", event.DeviceID),
		slog.String("zone", event.Zone),
		slog.String("event_id", event.EventID),
	)
	return nil
}

func (s *Service) toDLQ(ctx context.Context, msg repository.MQTTMessage, reason string) error {
	dlq := domain.DLQEvent{
		Envelope: domain.Envelope{
			EventID:       uuid.NewString(),
			EventType:     "telemetry.dlq",
			SchemaVersion: 1,
			OccurredAt:    time.Now().UTC(),
		},
		SourceTopic: msg.Topic,
		Reason:      reason,
		RawPayload:  string(msg.Payload),
	}

	metrics.KafkaPublished.WithLabelValues("raw.telemetry", "ok").Inc()

	if err := s.publisher.PublishDLQ(ctx, dlq); err != nil {
		return err
	}
	s.log.Warn("message sent to dlq",
		slog.String("topic", msg.Topic),
		slog.String("reason", reason),
	)

	return nil
}
