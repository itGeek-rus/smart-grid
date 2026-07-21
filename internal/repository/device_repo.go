package repository

import (
	"context"

	"github.com/itGeek-rus/smart-grid.git/internal/domain"
)

type DeviceRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Device, error)
	GetByExternalID(ctx context.Context, externalID string) (*domain.Device, error)
	ListByZone(ctx context.Context, zone string) ([]domain.Device, error)
	Upsert(ctx context.Context, device domain.Device) error
}
