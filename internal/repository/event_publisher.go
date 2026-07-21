package repository

import (
	"context"

	"github.com/itGeek-rus/smart-grid.git/internal/domain"
)

type EventPublisher interface {
	PublishRawTelemetry(ctx context.Context, event domain.RawTelemetryEvent) error
	PublishProcessed(ctx context.Context, event domain.ProcessedTelemetryEvent) error
	PublishAlert(ctx context.Context, event domain.AlertEvent) error
	PublishCommand(ctx context.Context, event domain.DeviceCommandEvent) error
	PublishDLQ(ctx context.Context, event domain.DLQEvent) error
}
