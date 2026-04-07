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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
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
		})
	}
}

func TestBikeHandler_ListMarketplace(t *testing.T) {
	tests := []struct {
		name           string
		mockBehavior   func(svc *MockBikeService)
		expectedStatus int
	}{
		{
			name: "Success 200 - returns bikes",
			mockBehavior: func(svc *MockBikeService) {
				svc.On("ListMarketplace", mock.Anything).Return([]domain.Bike{
					{ID: uuid.New(), Status: domain.StatusForSale},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Success 200 - empty list",
			mockBehavior: func(svc *MockBikeService) {
				svc.On("ListMarketplace", mock.Anything).Return([]domain.Bike{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Error 500 - service failure",
			mockBehavior: func(svc *MockBikeService) {
				svc.On("ListMarketplace", mock.Anything).Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/bikes", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockSvc := new(MockBikeService)
			tt.mockBehavior(mockSvc)
			h := &BikeHandler{service: mockSvc}

			if assert.NoError(t, h.ListMarketplace(c)) {
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestBikeHandler_ListMyBikes(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		mockBehavior   func(svc *MockBikeService)
		expectedStatus int
	}{
		{
			name: "Success 200",
			mockBehavior: func(svc *MockBikeService) {
				svc.On("ListMyBikes", mock.Anything, userID).Return([]domain.Bike{
					{ID: uuid.New(), CurrentOwnerID: userID},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Error 500 - service failure",
			mockBehavior: func(svc *MockBikeService) {
				svc.On("ListMyBikes", mock.Anything, userID).Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/my/bikes", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user", &auth.UserClaims{UserID: userID.String(), Role: "user"})

			mockSvc := new(MockBikeService)
			tt.mockBehavior(mockSvc)
			h := &BikeHandler{service: mockSvc}

			if assert.NoError(t, h.ListMyBikes(c)) {
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestBikeHandler_ListAdmin(t *testing.T) {
	adminID := uuid.New().String()

	tests := []struct {
		name           string
		role           string
		mockBehavior   func(svc *MockBikeService)
		expectedStatus int
	}{
		{
			name: "Success 200 - admin user",
			role: "admin",
			mockBehavior: func(svc *MockBikeService) {
				svc.On("ListAdmin", mock.Anything).Return([]domain.Bike{
					{ID: uuid.New()},
					{ID: uuid.New()},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Error 403 - non-admin user",
			role:           "user",
			mockBehavior:   func(svc *MockBikeService) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Error 500 - service failure for admin",
			role: "admin",
			mockBehavior: func(svc *MockBikeService) {
				svc.On("ListAdmin", mock.Anything).Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/admin/bikes", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user", &auth.UserClaims{UserID: adminID, Role: tt.role})

			mockSvc := new(MockBikeService)
			tt.mockBehavior(mockSvc)
			h := &BikeHandler{service: mockSvc}

			if assert.NoError(t, h.ListAdmin(c)) {
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestBikeHandler_RegisterBike_InvalidUserID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/bikes", strings.NewReader(`{"make_model":"Trek"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// Set a non-UUID user ID to trigger the parse error
	c.Set("user", &auth.UserClaims{UserID: "not-a-valid-uuid", Role: "user"})

	mockSvc := new(MockBikeService)
	h := &BikeHandler{service: mockSvc}

	if assert.NoError(t, h.RegisterBike(c)) {
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	}
}