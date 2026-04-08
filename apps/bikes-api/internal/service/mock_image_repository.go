package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/velotrace/bikes-api/internal/domain"
)

// MockImageRepository is a mock of the ImageRepository interface
type MockImageRepository struct {
	mock.Mock
}

func (m *MockImageRepository) GetImageCount(ctx context.Context, bikeID uuid.UUID) (int, error) {
	callArgs := m.Called(ctx, bikeID)
	return callArgs.Int(0), callArgs.Error(1)
}

func (m *MockImageRepository) CreateImage(ctx context.Context, img *domain.BikeImage) error {
	callArgs := m.Called(ctx, img)
	return callArgs.Error(0)
}
