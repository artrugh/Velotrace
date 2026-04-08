package service

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/velotrace/identity-api/internal/domain"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) UpsertByGoogleID(ctx context.Context, googleID, email, displayName string) (*domain.User, error) {
	args := m.Called(ctx, googleID, email, displayName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
