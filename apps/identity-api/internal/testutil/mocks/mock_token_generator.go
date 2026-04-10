package mocks

import (
	"github.com/stretchr/testify/mock"
	"velotrace.local/auth"
)

type MockTokenGenerator struct {
	mock.Mock
}

func (m *MockTokenGenerator) GenerateToken(claims auth.UserClaims) (string, error) {
	args := m.Called(claims)
	return args.String(0), args.Error(1)
}
