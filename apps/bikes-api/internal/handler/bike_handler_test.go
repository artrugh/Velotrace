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
	"github.com/velotrace/bikes-api/internal/service"
	"velotrace.local/auth"
)

// MockBikeService is a mock of the service layer
type MockBikeService struct {
	mock.Mock
}

func (m *MockBikeService) ListMarketplace(ctx context.Context, limit, offset int) ([]domain.Bike, int, int, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, 0, args.Error(3)
	}
	return args.Get(0).([]domain.Bike), args.Int(1), args.Int(2), args.Error(3)
}

func (m *MockBikeService) ListMyBikes(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Bike, int, int, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, 0, args.Error(3)
	}
	return args.Get(0).([]domain.Bike), args.Int(1), args.Int(2), args.Error(3)
}

func (m *MockBikeService) ListAdmin(ctx context.Context, limit, offset int) ([]domain.Bike, int, int, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, 0, args.Error(3)
	}
	return args.Get(0).([]domain.Bike), args.Int(1), args.Int(2), args.Error(3)
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
				svc.On("GetBike", mock.Anything, bikeID, userID, "user").Return(nil, errors.New("bike not found"))
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
				svc.On("RegisterBike", mock.Anything, mock.Anything).Return(service.ErrSerialNumberExists)
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
			payload:        `{"make_model":"Trek"}`,
			mockBehavior:   func(svc *MockBikeService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
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
	if req, ok := i.(*RegisterBikeRequest); ok {
		if req.MakeModel == "" || req.SerialNumber == "" {
			return errors.New("missing fields")
		}
	}
	return nil
}

func TestBikeHandler_ListMarketplace(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockBehavior   func(svc *MockBikeService)
		expectedStatus int
		expectedLimit  int
		expectedOffset int
		expectedTotal  int
	}{
		{
			name:        "Success with default pagination",
			queryParams: "",
			mockBehavior: func(svc *MockBikeService) {
				bikes := []domain.Bike{{ID: uuid.New(), MakeModel: "Trek"}}
				svc.On("ListMarketplace", mock.Anything, 1000, 0).Return(bikes, 1, 1000, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLimit:  1000,
			expectedOffset: 0,
			expectedTotal:  1,
		},
		{
			name:        "Success with custom limit and offset",
			queryParams: "?limit=10&offset=5",
			mockBehavior: func(svc *MockBikeService) {
				bikes := []domain.Bike{{ID: uuid.New(), MakeModel: "Trek"}}
				svc.On("ListMarketplace", mock.Anything, 10, 5).Return(bikes, 100, 10, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLimit:  10,
			expectedOffset: 5,
			expectedTotal:  100,
		},
		{
			name:           "Error 400 - Invalid limit",
			queryParams:    "?limit=invalid",
			mockBehavior:   func(svc *MockBikeService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error 400 - Invalid offset",
			queryParams:    "?offset=invalid",
			mockBehavior:   func(svc *MockBikeService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error 400 - Negative limit",
			queryParams:    "?limit=-1",
			mockBehavior:   func(svc *MockBikeService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error 400 - Negative offset",
			queryParams:    "?offset=-1",
			mockBehavior:   func(svc *MockBikeService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/bikes"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockSvc := new(MockBikeService)
			tt.mockBehavior(mockSvc)
			h := &BikeHandler{service: mockSvc}

			if assert.NoError(t, h.ListMarketplace(c)) {
				assert.Equal(t, tt.expectedStatus, rec.Code)
				if tt.expectedStatus == http.StatusOK {
					var resp BikeListResponse
					assert.NoError(t, echo.NewHTTPError(http.StatusOK).SetInternal(nil))
					// Verify response structure contains expected limit
					assert.Contains(t, rec.Body.String(), "\"limit\":")
				}
			}
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestBikeHandler_ListMyBikes(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		queryParams    string
		mockBehavior   func(svc *MockBikeService)
		expectedStatus int
		expectedLimit  int
		expectedOffset int
		expectedTotal  int
	}{
		{
			name:        "Success with default pagination",
			queryParams: "",
			mockBehavior: func(svc *MockBikeService) {
				bikes := []domain.Bike{{ID: uuid.New(), MakeModel: "Trek", CurrentOwnerID: userID}}
				svc.On("ListMyBikes", mock.Anything, userID, 1000, 0).Return(bikes, 1, 1000, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLimit:  1000,
			expectedOffset: 0,
			expectedTotal:  1,
		},
		{
			name:        "Success with custom limit and offset",
			queryParams: "?limit=20&offset=10",
			mockBehavior: func(svc *MockBikeService) {
				bikes := []domain.Bike{{ID: uuid.New(), MakeModel: "Trek", CurrentOwnerID: userID}}
				svc.On("ListMyBikes", mock.Anything, userID, 20, 10).Return(bikes, 50, 20, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLimit:  20,
			expectedOffset: 10,
			expectedTotal:  50,
		},
		{
			name:           "Error 400 - Invalid limit",
			queryParams:    "?limit=invalid",
			mockBehavior:   func(svc *MockBikeService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error 400 - Invalid offset",
			queryParams:    "?offset=invalid",
			mockBehavior:   func(svc *MockBikeService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/my/bikes"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user", &auth.UserClaims{UserID: userID.String(), Role: "user"})

			mockSvc := new(MockBikeService)
			tt.mockBehavior(mockSvc)
			h := &BikeHandler{service: mockSvc}

			if assert.NoError(t, h.ListMyBikes(c)) {
				assert.Equal(t, tt.expectedStatus, rec.Code)
				if tt.expectedStatus == http.StatusOK {
					assert.Contains(t, rec.Body.String(), "\"limit\":")
				}
			}
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestBikeHandler_ListAdmin(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		userRole       string
		mockBehavior   func(svc *MockBikeService)
		expectedStatus int
		expectedLimit  int
		expectedOffset int
		expectedTotal  int
	}{
		{
			name:        "Success with default pagination",
			queryParams: "",
			userRole:    "admin",
			mockBehavior: func(svc *MockBikeService) {
				bikes := []domain.Bike{{ID: uuid.New(), MakeModel: "Trek"}}
				svc.On("ListAdmin", mock.Anything, 1000, 0).Return(bikes, 1, 1000, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLimit:  1000,
			expectedOffset: 0,
			expectedTotal:  1,
		},
		{
			name:        "Success with custom limit and offset",
			queryParams: "?limit=50&offset=25",
			userRole:    "admin",
			mockBehavior: func(svc *MockBikeService) {
				bikes := []domain.Bike{{ID: uuid.New(), MakeModel: "Trek"}}
				svc.On("ListAdmin", mock.Anything, 50, 25).Return(bikes, 200, 50, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLimit:  50,
			expectedOffset: 25,
			expectedTotal:  200,
		},
		{
			name:           "Error 403 - Not admin",
			queryParams:    "",
			userRole:       "user",
			mockBehavior:   func(svc *MockBikeService) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Error 400 - Invalid limit",
			queryParams:    "?limit=invalid",
			userRole:       "admin",
			mockBehavior:   func(svc *MockBikeService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error 400 - Invalid offset",
			queryParams:    "?offset=invalid",
			userRole:       "admin",
			mockBehavior:   func(svc *MockBikeService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/admin/bikes"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user", &auth.UserClaims{UserID: uuid.New().String(), Role: tt.userRole})

			mockSvc := new(MockBikeService)
			tt.mockBehavior(mockSvc)
			h := &BikeHandler{service: mockSvc}

			if assert.NoError(t, h.ListAdmin(c)) {
				assert.Equal(t, tt.expectedStatus, rec.Code)
				if tt.expectedStatus == http.StatusOK {
					assert.Contains(t, rec.Body.String(), "\"limit\":")
				}
			}
			mockSvc.AssertExpectations(t)
		})
	}
}