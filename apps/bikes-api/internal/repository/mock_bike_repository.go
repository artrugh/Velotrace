package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/velotrace/bikes-api/internal/domain"
)

// MockBikeRepository is a mock of the BikeRepository interface
type MockBikeRepository struct {
	mock.Mock
}

func (m *MockBikeRepository) GetAll(ctx context.Context, filter string, args []interface{}) ([]domain.Bike, error) {
	callArgs := m.Called(ctx, filter, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).([]domain.Bike), callArgs.Error(1)
}

func (m *MockBikeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bike, error) {
	callArgs := m.Called(ctx, id)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(*domain.Bike), callArgs.Error(1)
}

func (m *MockBikeRepository) Create(ctx context.Context, bike *domain.Bike) error {
	callArgs := m.Called(ctx, bike)
	return callArgs.Error(0)
}

func (m *MockBikeRepository) GetBikeImages(ctx context.Context, bikeID uuid.UUID) ([]domain.BikeImage, error) {
	callArgs := m.Called(ctx, bikeID)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).([]domain.BikeImage), callArgs.Error(1)
}
