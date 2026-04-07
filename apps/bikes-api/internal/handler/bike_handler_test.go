package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/velotrace/bikes-api/internal/domain"
	"velotrace.local/auth"
)

// MockBikeService is a mock of the service layer
type MockBikeService struct {
	mock.Mock
}

func (m *MockBikeService) ListMarketplace(ctx context.Context) ([]domain.Bike, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Bike), args.Error(1)
}

func (m *MockBikeService) ListMyBikes(ctx context.Context, userID uuid.UUID) ([]domain.Bike, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Bike), args.Error(1)
}

func (m *MockBikeService) ListAdmin(ctx context.Context) ([]domain.Bike, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Bike), args.Error(1)
}

func (m *MockBikeService) GetBike(ctx context.Context, id uuid.UUID, userID string, role string) (*domain.Bike, error) {
	args := m.Called(ctx, id, userID, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Bike), args.Error(1)
}

func (m *MockBikeService) RegisterBike(ctx context.Context, bike *domain.Bike) error {
	args := m.Called(ctx, bike)
	return args.Error(0)
}

func TestBikeHandler_GetBike(t *testing.T) {
	bikeID := uuid.New()
	userID := uuid.New().String()

	tests := []struct {
		name           string
		bikeID         string
		userClaims     *auth.UserClaims
		mockBehavior   func(svc *MockBikeService)
		expectedStatus int
	}{
		{
			name:       "Success 200",
			bikeID:     bikeID.String(),
			userClaims: &auth.UserClaims{UserID: userID, Role: "user"},
			mockBehavior: func(svc *MockBikeService) {
				svc.On("GetBike", mock.Anything, bikeID, userID, "user").Return(&domain.Bike{ID: bikeID}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "Error 404 - Not Found",
			bikeID:     bikeID.String(),
			userClaims: &auth.UserClaims{UserID: userID, Role: "user"},
			mockBehavior: func(svc *MockBikeService) {
				svc.On("GetBike", mock.Anything, bikeID, userID, "user").Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Error 400 - Invalid UUID",
			bikeID:         "invalid-uuid",
			userClaims:     &auth.UserClaims{UserID: userID, Role: "user"},
			mockBehavior:   func(svc *MockBikeService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/bikes/"+tt.bikeID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.bikeID)
			c.Set("user", tt.userClaims)

			mockSvc := new(MockBikeService)
			tt.mockBehavior(mockSvc)
			h := &BikeHandler{service: mockSvc}

			if assert.NoError(t, h.GetBike(c)) {
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestBikeHandler_RegisterBike(t *testing.T) {
	userID := uuid.New().String()
	validPayload := `{"make_model":"Trek","serial_number":"123"}`

	tests := []struct {
		name           string
		payload        string
		mockBehavior   func(svc *MockBikeService)
		expectedStatus int
	}{
		{
			name:    "Success 201",
			payload: validPayload,
			mockBehavior: func(svc *MockBikeService) {
				svc.On("RegisterBike", mock.Anything, mock.MatchedBy(func(b *domain.Bike) bool {
					return b.MakeModel == "Trek"
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:    "Error 409 - Conflict",
			payload: validPayload,
			mockBehavior: func(svc *MockBikeService) {
				svc.On("RegisterBike", mock.Anything, mock.Anything).Return(errors.New("serial number already registered"))
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "Error 400 - Bad JSON",
			payload:        `{invalid}`,
			mockBehavior:   func(svc *MockBikeService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error 400 - Missing Field",
			payload:        `{"make_model":"Trek"}`, // SerialNumber missing
			mockBehavior:   func(svc *MockBikeService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			// Mock validator
			cv := &MockValidator{}
			e.Validator = cv

			req := httptest.NewRequest(http.MethodPost, "/bikes", strings.NewReader(tt.payload))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user", &auth.UserClaims{UserID: userID, Role: "user"})

			mockSvc := new(MockBikeService)
			tt.mockBehavior(mockSvc)
			h := &BikeHandler{service: mockSvc}

			if assert.NoError(t, h.RegisterBike(c)) {
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}
			mockSvc.AssertExpectations(t)
		})
	}
}

type MockValidator struct{}

func (v *MockValidator) Validate(i interface{}) error {
	// Simple validation for testing
	if req, ok := i.(*RegisterBikeRequest); ok {
		if req.MakeModel == "" || req.SerialNumber == "" {
			return errors.New("missing fields")
		}
	}
	return nil
}
