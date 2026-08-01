package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/itGeek-rus/smart-grid.git/internal/domain"
	"github.com/itGeek-rus/smart-grid.git/internal/repository"
)

type Service struct {
	telemetry repository.TelemetryRepository
	alerts    repository.AlertRepository
	cache     repository.TelemetryCache
	publisher repository.EventPublisher
	log       *slog.Logger
	window    time.Duration
}

func NewService(
	telemetry repository.TelemetryRepository,
	alerts repository.AlertRepository,
	cache repository.TelemetryCache,
	publisher repository.EventPublisher,
	log *slog.Logger,
) *Service {
	return &Service{
		telemetry: telemetry,
		alerts:    alerts,
		cache:     cache,
		publisher: publisher,
		log:       log,
		window:    5 * time.Minute,
	}
}

func (s *Service) HandleRawMessage(ctx context.Context, key, value []byte) error {
	var event domain.RawTelemetryEvent
	if err := json.Unmarshal(value, &event); err != nil {
		return s.toDLQ(ctx, key, value, fmt.Sprintf("json: %v", err))
	}
	if event.DeviceID == "" || event.MeasuredAt.IsZero() {
		return s.toDLQ(ctx, key, value, "missing device_id or measured_at")
	}

	t := domain.Telemetry{
		ID:         event.EventID,
		DeviceID:   event.DeviceID,
		Zone:       event.Zone,
		Voltage:    event.Voltage,
		Current:    event.Current,
		Power:      event.Power,
		Frequency:  event.Frequency,
		MeasuredAt: event.MeasuredAt,
		ReceivedAt: event.ReceivedAt,
		Payload:    event.Payload,
	}

	if err := s.telemetry.Insert(ctx, t); err != nil {
		return err
	}
	if err := s.cache.AddSample(ctx, t.DeviceID, t.Power, t.MeasuredAt); err != nil {
		return err
	}
	_ = s.cache.SetLastTelemetry(ctx, t)

	avgPower, maxPower, samples, err := s.cache.WindowStats(ctx, t.DeviceID, s.window)
	if err != nil {
		return err
	}

	det := DetectAnomaly(t.Voltage, t.Power, avgPower, samples)

	processed := domain.ProcessedTelemetryEvent{
		Envelope: domain.Envelope{
			EventID:       uuid.NewString(),
			EventType:     domain.EventTypeProcessedTelemetry,
			SchemaVersion: 1,
			OccurredAt:    time.Now().UTC(),
			TraceID:       event.TraceID,
		},
		DeviceID:     t.DeviceID,
		Zone:         t.Zone,
		Power:        t.Power,
		AnomalyScore: det.AnomalyScore,
		ProcessedAt:  time.Now().UTC(),
	}
	if err := s.publisher.PublishProcessed(ctx, processed); err != nil {
		return err
	}

	agg := BuildFiveMinAggregate(t.DeviceID, t.Voltage, t.Current,
		avgPower, maxPower, samples, t.MeasuredAt)
	_ = s.telemetry.InsertAggregate(ctx, agg)

	if det.IsAnomaly {
		alert := domain.Alert{
			ID:         uuid.NewString(),
			DeviceID:   t.DeviceID,
			Type:       det.AlertType,
			Severity:   domain.AlertSeverityWarning,
			Status:     domain.AlertStatusOpen,
			Message:    det.Message,
			Score:      det.AnomalyScore,
			DetectedAt: time.Now().UTC(),
			CreatedAt:  time.Now().UTC(),
		}
		if det.AnomalyScore >= 0.9 {
			alert.Severity = domain.AlertSeverityCritical
		}
		if err := s.alerts.Create(ctx, alert); err != nil {
			return err
		}
		if err := s.publisher.PublishAlert(ctx, domain.AlertEvent{
			Envelope: domain.Envelope{
				EventID:       uuid.NewString(),
				EventType:     domain.EventTypeAlertRaised,
				SchemaVersion: 1,
				OccurredAt:    time.Now().UTC(),
			},
			AlertID:    alert.ID,
			DeviceID:   alert.DeviceID,
			Type:       alert.Type,
			Severity:   alert.Severity,
			Message:    alert.Message,
			Score:      alert.Score,
			DetectedAt: alert.DetectedAt,
		}); err != nil {
			return err
		}
	}

	s.log.Info("telemetry processed",
		slog.String("device_id", t.DeviceID),
		slog.Float64("anomaly_score", det.AnomalyScore),
	)

	return nil
}

func (s *Service) toDLQ(ctx context.Context, key, value []byte, reason string) error {
	err := s.publisher.PublishDLQ(ctx, domain.DLQEvent{
		Envelope: domain.Envelope{
			EventID:       uuid.NewString(),
			EventType:     "telemetry.dlq",
			SchemaVersion: 1,
			OccurredAt:    time.Now().UTC(),
		},
		SourceTopic: "raw.telemetry",
		Reason:      reason,
		RawPayload:  string(value),
	})
	if err != nil {
		return err
	}
	s.log.Warn("raw message sent to dlq",
		slog.String("reason", reason),
		slog.String("key", string(key)),
	)
	return nil
}
