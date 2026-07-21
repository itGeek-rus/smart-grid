package repository

import (
	"context"

	"github.com/itGeek-rus/smart-grid.git/internal/domain"
)

type AlertRepository interface {
	Create(ctx context.Context, alert domain.Alert) error
	GetByID(ctx context.Context, id string) (*domain.Alert, error)
	ListOpenByDevice(ctx context.Context, deviceID string) ([]domain.Alert, error)
	UpdateStatus(ctx context.Context, id string, status domain.AlertStatus) error
}
