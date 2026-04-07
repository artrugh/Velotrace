package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/velotrace/bikes-api/internal/domain"
)

type ImageRepository interface {
	GetImageCount(ctx context.Context, bikeID uuid.UUID) (int, error)
	CreateImage(ctx context.Context, img *domain.BikeImage) error
}
