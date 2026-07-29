package usecase

import (
	"context"

	"github.com/itGeek-rus/smart-grid.git/internal/repository"
)

type IngestionUseCase struct {
	handler func(ctx context.Context, msg repository.MQTTMessage) error
}

func NewIngestionUseCase(handler func(ctx context.Context, msg repository.MQTTMessage) error) *IngestionUseCase {
	return &IngestionUseCase{handler: handler}
}

func (uc *IngestionUseCase) Handle(ctx context.Context, msg repository.MQTTMessage) error {
	return uc.handler(ctx, msg)
}
