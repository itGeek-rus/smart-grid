package repository

import (
	"context"
	"time"

	"github.com/itGeek-rus/smart-grid.git/internal/domain"
)

type TelemetryRepository interface {
	Insert(ctx context.Context, t domain.Telemetry) error
	InsertBatch(ctx context.Context, items []domain.Telemetry) error
	ListByDevice(ctx context.Context, deviceID string, from, to time.Time, limit int) ([]domain.Telemetry, error)
	InsertAggregate(ctx context.Context, agg domain.Telemetry) error
}
