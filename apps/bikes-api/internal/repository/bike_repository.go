package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/velotrace/bikes-api/internal/domain"
)

type BikeRepository interface {
	GetAll(ctx context.Context, filter string, args []interface{}) ([]domain.Bike, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Bike, error)
	Create(ctx context.Context, bike *domain.Bike) error
	GetBikeImages(ctx context.Context, bikeID uuid.UUID) ([]domain.BikeImage, error)
}
