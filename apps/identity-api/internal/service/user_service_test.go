package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/velotrace/identity-api/internal/domain"
	"github.com/velotrace/identity-api/internal/repository"
	"google.golang.org/api/idtoken"
)

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

func generateTestRSAPrivateKey(t *testing.T) string {
	t.Helper()
	// Generate a temporary RSA key pair at runtime
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	// PEM-encode the private key
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return string(privateKeyPEM)
}

func TestUserService_AuthGoogle(t *testing.T) {
	ctx := context.Background()
	testPrivKey := generateTestRSAPrivateKey(t)

	tests := []struct {
		name          string
		credential    string
		setupEnv      func(t *testing.T)
		mockValidator func(m *MockTokenValidator)
		mockRepo      func(m *repository.MockUserRepository)
		wantErr       bool
		expectedErr   error
	}{
		{
			name:       "Success",
			credential: "valid-token",
			setupEnv: func(t *testing.T) {
				t.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
				t.Setenv("JWT_PRIVATE_KEY", testPrivKey)
			},
			mockValidator: func(m *MockTokenValidator) {
				m.On("Validate", ctx, "valid-token", "test-client-id").Return(&idtoken.Payload{
					Subject: "google-id-123",
					Claims: map[string]interface{}{
						"email": "user@example.com",
						"name":  "John Doe",
					},
				}, nil)
			},
			mockRepo: func(m *repository.MockUserRepository) {
				m.On("UpsertByGoogleID", ctx, "google-id-123", "user@example.com", "John Doe").Return(&domain.User{
					ID:    uuid.New(),
					Email: "user@example.com",
					Role:  "user",
				}, nil)
			},
			wantErr: false,
		},
		{
			name:       "Missing Google Client ID",
			credential: "any-token",
			setupEnv: func(t *testing.T) {
				t.Setenv("GOOGLE_CLIENT_ID", "")
			},
			mockValidator: func(m *MockTokenValidator) {},
			mockRepo:      func(m *repository.MockUserRepository) {},
			wantErr:       true,
			expectedErr:   ErrMissingClientID,
		},
		{
			name:       "Invalid Token",
			credential: "invalid-token",
			setupEnv: func(t *testing.T) {
				t.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
				t.Setenv("JWT_PRIVATE_KEY", testPrivKey)
			},
			mockValidator: func(m *MockTokenValidator) {
				m.On("Validate", ctx, "invalid-token", "test-client-id").Return(nil, errors.New("invalid token"))
			},
			mockRepo:    func(m *repository.MockUserRepository) {},
			wantErr:     true,
			expectedErr: ErrInvalidGoogleToken,
		},
		{
			name:       "Email Claim Missing",
			credential: "token-no-email",
			setupEnv: func(t *testing.T) {
				t.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
				t.Setenv("JWT_PRIVATE_KEY", testPrivKey)
			},
			mockValidator: func(m *MockTokenValidator) {
				m.On("Validate", ctx, "token-no-email", "test-client-id").Return(&idtoken.Payload{
					Subject: "google-id-123",
					Claims: map[string]interface{}{
						"name": "John Doe",
					},
				}, nil)
			},
			mockRepo:    func(m *repository.MockUserRepository) {},
			wantErr:     true,
			expectedErr: ErrEmailClaimMissing,
		},
		{
			name:       "Failed to Generate Token",
			credential: "valid-token",
			setupEnv: func(t *testing.T) {
				t.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
				t.Setenv("JWT_PRIVATE_KEY", "invalid-key") // This will cause GenerateToken to fail
			},
			mockValidator: func(m *MockTokenValidator) {
				m.On("Validate", ctx, "valid-token", "test-client-id").Return(&idtoken.Payload{
					Subject: "google-id-123",
					Claims: map[string]interface{}{
						"email": "user@example.com",
						"name":  "John Doe",
					},
				}, nil)
			},
			mockRepo: func(m *repository.MockUserRepository) {
				m.On("UpsertByGoogleID", ctx, "google-id-123", "user@example.com", "John Doe").Return(&domain.User{
					ID:    uuid.New(),
					Email: "user@example.com",
					Role:  "user",
				}, nil)
			},
			wantErr:     true,
			expectedErr: ErrFailedToGenerateToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv(t)
			mockValidator := new(MockTokenValidator)
			mockRepo := new(repository.MockUserRepository)
			tt.mockValidator(mockValidator)
			tt.mockRepo(mockRepo)

			s := &userService{
				repo:      mockRepo,
				validator: mockValidator,
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
		})
	}
}
