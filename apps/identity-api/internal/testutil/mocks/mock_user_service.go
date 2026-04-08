package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/velotrace/identity-api/internal/domain"
	"google.golang.org/api/idtoken"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) AuthGoogle(ctx context.Context, credential string) (*domain.User, string, error) {
	args := m.Called(ctx, credential)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*domain.User), args.String(1), args.Error(2)
}

type MockTokenValidator struct {
	mock.Mock
}

func (m *MockTokenValidator) Validate(ctx context.Context, idToken string, audience string) (*idtoken.Payload, error) {
	args := m.Called(ctx, idToken, audience)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*idtoken.Payload), args.Error(1)
}
