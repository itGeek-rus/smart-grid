package repository

import (
	"context"
	"time"

	"github.com/itGeek-rus/smart-grid.git/internal/domain"
)

type TelemetryCache interface {
	AddSample(ctx context.Context, deviceID string, power float64, at time.Time) error
	WindowStats(ctx context.Context, deviceID string, window time.Duration) (avgPower, maxPower float64, samples int64, err error)
	SetLastTelemetry(ctx context.Context, t domain.Telemetry) error
	GetLastTelemetry(ctx context.Context, deviceID string) (*domain.Telemetry, error)
}
