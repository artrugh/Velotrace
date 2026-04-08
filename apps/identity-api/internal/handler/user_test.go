package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/velotrace/identity-api/internal/domain"
	"github.com/velotrace/identity-api/internal/service"
)

func TestUserHandler_AuthGoogle(t *testing.T) {
	tests := []struct {
		name           string
		payload        string
		mockBehavior   func(svc *service.MockUserService)
		expectedStatus int
	}{
		{
			name:    "Success 200",
			payload: `{"credential":"valid-token"}`,
			mockBehavior: func(svc *service.MockUserService) {
				svc.On("AuthGoogle", mock.Anything, "valid-token").Return(&domain.User{ID: uuid.New(), Email: "test@example.com"}, "fake-jwt", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Error 401 - Invalid Token",
			payload: `{"credential":"invalid-token"}`,
			mockBehavior: func(svc *service.MockUserService) {
				svc.On("AuthGoogle", mock.Anything, "invalid-token").Return(nil, "", service.ErrInvalidGoogleToken)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Error 400 - Missing Credential",
			payload:        `{"credential":""}`,
			mockBehavior:   func(svc *service.MockUserService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Error 401 - Email Claim Missing",
			payload: `{"credential":"token-no-email"}`,
			mockBehavior: func(svc *service.MockUserService) {
				svc.On("AuthGoogle", mock.Anything, "token-no-email").Return(nil, "", service.ErrEmailClaimMissing)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:    "Error 500 - Internal Server Error",
			payload: `{"credential":"valid-token"}`,
			mockBehavior: func(svc *service.MockUserService) {
				svc.On("AuthGoogle", mock.Anything, "valid-token").Return(nil, "", errors.New("database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/auth/google", strings.NewReader(tt.payload))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockSvc := new(service.MockUserService)
			tt.mockBehavior(mockSvc)
			h := NewUserHandler(mockSvc)

			if assert.NoError(t, h.AuthGoogle(c)) {
				assert.Equal(t, tt.expectedStatus, rec.Code)

				// For success cases, verify response body
				if tt.expectedStatus == http.StatusOK {
					body := rec.Body.String()
					assert.Contains(t, body, `"user":{`)
					assert.Contains(t, body, `"token":"fake-jwt"`)
					assert.Contains(t, body, `"email":"test@example.com"`)
				}
			}
			mockSvc.AssertExpectations(t)
		})
	}
}
