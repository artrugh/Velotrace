package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/velotrace/identity-api/internal/domain"
	"github.com/velotrace/identity-api/internal/testutil/mocks"
	"google.golang.org/api/idtoken"
)

func TestUserService_AuthGoogle(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		credential    string
		setupEnv      func(t *testing.T)
		mockValidator func(m *mocks.MockTokenValidator)
		mockRepo      func(m *mocks.MockUserRepository)
		mockGenerator func(m *mocks.MockTokenGenerator)
		wantErr       bool
		expectedErr   error
	}{
		{
			name:       "Success",
			credential: "valid-token",
			setupEnv: func(t *testing.T) {
				t.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
			},
			mockValidator: func(m *mocks.MockTokenValidator) {
				m.On("Validate", ctx, "valid-token", "test-client-id").Return(&idtoken.Payload{
					Subject: "google-id-123",
					Claims: map[string]interface{}{
						"email": "user@example.com",
						"name":  "John Doe",
					},
				}, nil)
			},
			mockRepo: func(m *mocks.MockUserRepository) {
				m.On("UpsertByGoogleID", ctx, "google-id-123", "user@example.com", "John Doe").Return(&domain.User{
					ID:    uuid.New(),
					Email: "user@example.com",
					Role:  "user",
				}, nil)
			},
			mockGenerator: func(m *mocks.MockTokenGenerator) {
				m.On("GenerateToken", mock.Anything).Return("session-token", nil)
			},
			wantErr: false,
		},
		{
			name:       "Missing Google Client ID",
			credential: "any-token",
			setupEnv: func(t *testing.T) {
				t.Setenv("GOOGLE_CLIENT_ID", "")
			},
			mockValidator: func(m *mocks.MockTokenValidator) {},
			mockRepo:      func(m *mocks.MockUserRepository) {},
			mockGenerator: func(m *mocks.MockTokenGenerator) {},
			wantErr:       true,
			expectedErr:   ErrMissingClientID,
		},
		{
			name:       "Invalid Token",
			credential: "invalid-token",
			setupEnv: func(t *testing.T) {
				t.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
			},
			mockValidator: func(m *mocks.MockTokenValidator) {
				m.On("Validate", ctx, "invalid-token", "test-client-id").Return(nil, errors.New("invalid token"))
			},
			mockRepo:      func(m *mocks.MockUserRepository) {},
			mockGenerator: func(m *mocks.MockTokenGenerator) {},
			wantErr:       true,
			expectedErr:   ErrInvalidGoogleToken,
		},
		{
			name:       "Email Claim Missing",
			credential: "token-no-email",
			setupEnv: func(t *testing.T) {
				t.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
			},
			mockValidator: func(m *mocks.MockTokenValidator) {
				m.On("Validate", ctx, "token-no-email", "test-client-id").Return(&idtoken.Payload{
					Subject: "google-id-123",
					Claims: map[string]interface{}{
						"name": "John Doe",
					},
				}, nil)
			},
			mockRepo:      func(m *mocks.MockUserRepository) {},
			mockGenerator: func(m *mocks.MockTokenGenerator) {},
			wantErr:       true,
			expectedErr:   ErrEmailClaimMissing,
		},
		{
			name:       "Failed to Generate Token",
			credential: "valid-token",
			setupEnv: func(t *testing.T) {
				t.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
			},
			mockValidator: func(m *mocks.MockTokenValidator) {
				m.On("Validate", ctx, "valid-token", "test-client-id").Return(&idtoken.Payload{
					Subject: "google-id-123",
					Claims: map[string]interface{}{
						"email": "user@example.com",
						"name":  "John Doe",
					},
				}, nil)
			},
			mockRepo: func(m *mocks.MockUserRepository) {
				m.On("UpsertByGoogleID", ctx, "google-id-123", "user@example.com", "John Doe").Return(&domain.User{
					ID:    uuid.New(),
					Email: "user@example.com",
					Role:  "user",
				}, nil)
			},
			mockGenerator: func(m *mocks.MockTokenGenerator) {
				m.On("GenerateToken", mock.Anything).Return("", errors.New("generation failed"))
			},
			wantErr:     true,
			expectedErr: ErrFailedToGenerateToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv(t)
			mockValidator := new(mocks.MockTokenValidator)
			mockRepo := new(mocks.MockUserRepository)
			mockGenerator := new(mocks.MockTokenGenerator)

			tt.mockValidator(mockValidator)
			tt.mockRepo(mockRepo)
			tt.mockGenerator(mockGenerator)

			s := &userService{
				repo:        mockRepo,
				authManager: mockGenerator,
				validator:   mockValidator,
			}

			user, token, err := s.AuthGoogle(ctx, tt.credential)

			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr), "expected error %v, got %v", tt.expectedErr, err)
				assert.Nil(t, user)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, token)
			}

			mockValidator.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
			mockGenerator.AssertExpectations(t)
		})
	}
}
