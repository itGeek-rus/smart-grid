package usecase

import (
	"context"
	"time"

	"github.com/itGeek-rus/smart-grid.git/internal/domain"
	"github.com/itGeek-rus/smart-grid.git/internal/repository"
)

type APIUseCase struct {
	devices   repository.DeviceRepository
	telemetry repository.TelemetryRepository
	alerts    repository.AlertRepository
	cache     repository.TelemetryCache
	publisher repository.EventPublisher
}

func NewAPIUseCase(
	devices repository.DeviceRepository,
	telemetry repository.TelemetryRepository,
	alerts repository.AlertRepository,
	cache repository.TelemetryCache,
	publisher repository.EventPublisher,
) *APIUseCase {
	return &APIUseCase{
		devices:   devices,
		telemetry: telemetry,
		alerts:    alerts,
		cache:     cache,
		publisher: publisher,
	}
}

func (uc *APIUseCase) ListDevices(ctx context.Context, zone string) ([]domain.Device, error) {
	if zone == "" {
		return uc.devices.ListAll(ctx)
	}
	return uc.devices.ListByZone(ctx, zone)
}

func (uc *APIUseCase) GetDevice(ctx context.Context, id string) (*domain.Device, error) {
	return uc.devices.GetByID(ctx, id)
}

func (uc *APIUseCase) GetTelemetry(
	ctx context.Context,
	deviceID string,
	from, to time.Time,
	limit int,
) ([]domain.Telemetry, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	if to.IsZero() {
		to = time.Now().UTC()
	}
	if from.IsZero() {
		from = to.Add(-1 * time.Hour)
	}
	return uc.telemetry.ListByDevice(ctx, deviceID, from, to, limit)
}

func (uc *APIUseCase) GetLatestTelemetry(ctx context.Context, deviceID string) (*domain.Telemetry, error) {
	if t, err := uc.cache.GetLastTelemetry(ctx, deviceID); err == nil {
		return t, nil
	}
	items, err := uc.telemetry.ListByDevice(
		ctx, deviceID, time.Now().Add(-24*time.Hour),
		time.Now().UTC(), 1)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, domain.ErrNotFound
	}
	return &items[0], nil
}

func (uc *APIUseCase) ListOpenAlerts(ctx context.Context, deviceID string) ([]domain.Alert, error) {
	return uc.alerts.ListOpenByDevice(ctx, deviceID)
}

func (uc *APIUseCase) SendCommand(ctx context.Context, cmd domain.DeviceCommandEvent) error {
	return uc.publisher.PublishCommand(ctx, cmd)
}

func (uc *APIUseCase) ListAllDevices(ctx context.Context) ([]domain.Device, error) {
	return uc.devices.ListAll(ctx)
}
