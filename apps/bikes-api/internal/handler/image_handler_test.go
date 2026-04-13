package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/velotrace/bikes-api/internal/domain"
	"github.com/velotrace/bikes-api/internal/service"
	"github.com/velotrace/bikes-api/internal/testutil/mocks"
	"velotrace.local/auth"
)

// echoValidator wraps go-playground/validator for use with echo
type echoValidator struct {
	v *validator.Validate
}

func (ev *echoValidator) Validate(i interface{}) error {
	return ev.v.Struct(i)
}

func newTestEcho() *echo.Echo {
	e := echo.New()
	e.Validator = &echoValidator{v: validator.New()}
	return e
}

// buildImageService creates a real ImageService with mock repos (nil storage for ConfirmUpload tests)
func buildImageService(bikeRepo domain.BikeRepository, imageRepo domain.ImageRepository) *service.ImageService {
	return service.NewImageService(imageRepo, bikeRepo, nil)
}

func TestImageHandler_GetUploadURL_InvalidBikeID(t *testing.T) {
	e := newTestEcho()
	req := httptest.NewRequest(http.MethodPost, "/bikes/invalid-id/upload-url",
		strings.NewReader(`{"filename":"photo.jpg"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	mockBikeRepo := new(mocks.MockBikeRepository)
	mockImageRepo := new(mocks.MockImageRepository)
	h := &ImageHandler{service: buildImageService(mockBikeRepo, mockImageRepo)}

	if assert.NoError(t, h.GetUploadURL(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestImageHandler_GetUploadURL_InvalidBody(t *testing.T) {
	bikeID := uuid.New()

	e := newTestEcho()
	req := httptest.NewRequest(http.MethodPost, "/bikes/"+bikeID.String()+"/upload-url",
		strings.NewReader(`{not valid json}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(bikeID.String())

	mockBikeRepo := new(mocks.MockBikeRepository)
	mockImageRepo := new(mocks.MockImageRepository)
	h := &ImageHandler{service: buildImageService(mockBikeRepo, mockImageRepo)}

	if assert.NoError(t, h.GetUploadURL(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestImageHandler_GetUploadURL_MissingFilename(t *testing.T) {
	bikeID := uuid.New()

	e := newTestEcho()
	req := httptest.NewRequest(http.MethodPost, "/bikes/"+bikeID.String()+"/upload-url",
		strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(bikeID.String())

	mockBikeRepo := new(mocks.MockBikeRepository)
	mockImageRepo := new(mocks.MockImageRepository)
	h := &ImageHandler{service: buildImageService(mockBikeRepo, mockImageRepo)}

	if assert.NoError(t, h.GetUploadURL(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestImageHandler_ConfirmUpload(t *testing.T) {
	bikeID := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()

	tests := []struct {
		name           string
		bikeIDParam    string
		payload        string
		userClaims     *auth.UserClaims
		mockBikeRepo   func(repo *mocks.MockBikeRepository)
		mockImageRepo  func(repo *mocks.MockImageRepository)
		expectedStatus int
	}{
		{
			name:           "Error 400 - Invalid bike UUID",
			bikeIDParam:    "not-a-uuid",
			payload:        `{"object_key":"bikes/abc/photo.jpg"}`,
			userClaims:     &auth.UserClaims{UserID: ownerID.String(), Role: "user"},
			mockBikeRepo:   func(repo *mocks.MockBikeRepository) {},
			mockImageRepo:  func(repo *mocks.MockImageRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error 400 - Invalid JSON body",
			bikeIDParam:    bikeID.String(),
			payload:        `{invalid}`,
			userClaims:     &auth.UserClaims{UserID: ownerID.String(), Role: "user"},
			mockBikeRepo:   func(repo *mocks.MockBikeRepository) {},
			mockImageRepo:  func(repo *mocks.MockImageRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error 400 - Missing object_key",
			bikeIDParam:    bikeID.String(),
			payload:        `{}`,
			userClaims:     &auth.UserClaims{UserID: ownerID.String(), Role: "user"},
			mockBikeRepo:   func(repo *mocks.MockBikeRepository) {},
			mockImageRepo:  func(repo *mocks.MockImageRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Error 404 - Bike not found",
			bikeIDParam: bikeID.String(),
			payload:     `{"object_key":"bikes/abc/photo.jpg"}`,
			userClaims:  &auth.UserClaims{UserID: ownerID.String(), Role: "user"},
			mockBikeRepo: func(repo *mocks.MockBikeRepository) {
				repo.On("GetByID", mock.Anything, bikeID).Return(nil, service.ErrBikeNotFound)
			},
			mockImageRepo:  func(repo *mocks.MockImageRepository) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "Error 403 - Not the bike owner",
			bikeIDParam: bikeID.String(),
			payload:     `{"object_key":"bikes/abc/photo.jpg"}`,
			userClaims:  &auth.UserClaims{UserID: otherUserID.String(), Role: "user"},
			mockBikeRepo: func(repo *mocks.MockBikeRepository) {
				repo.On("GetByID", mock.Anything, bikeID).Return(&domain.Bike{
					ID: bikeID, CurrentOwnerID: ownerID,
				}, nil)
			},
			mockImageRepo:  func(repo *mocks.MockImageRepository) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:        "Success 201 - First image is primary",
			bikeIDParam: bikeID.String(),
			payload:     `{"object_key":"bikes/abc/photo.jpg"}`,
			userClaims:  &auth.UserClaims{UserID: ownerID.String(), Role: "user"},
			mockBikeRepo: func(repo *mocks.MockBikeRepository) {
				repo.On("GetByID", mock.Anything, bikeID).Return(&domain.Bike{
					ID: bikeID, CurrentOwnerID: ownerID,
				}, nil)
			},
			mockImageRepo: func(repo *mocks.MockImageRepository) {
				repo.On("GetImageCount", mock.Anything, bikeID).Return(0, nil)
				repo.On("CreateImage", mock.Anything, mock.MatchedBy(func(img *domain.BikeImage) bool {
					return img.IsPrimary == true
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:        "Error 500 - Image creation failure",
			bikeIDParam: bikeID.String(),
			payload:     `{"object_key":"bikes/abc/photo.jpg"}`,
			userClaims:  &auth.UserClaims{UserID: ownerID.String(), Role: "user"},
			mockBikeRepo: func(repo *mocks.MockBikeRepository) {
				repo.On("GetByID", mock.Anything, bikeID).Return(&domain.Bike{
					ID: bikeID, CurrentOwnerID: ownerID,
				}, nil)
			},
			mockImageRepo: func(repo *mocks.MockImageRepository) {
				repo.On("GetImageCount", mock.Anything, bikeID).Return(1, nil)
				repo.On("CreateImage", mock.Anything, mock.Anything).Return(errors.New("db insert failed"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			req := httptest.NewRequest(http.MethodPost, "/bikes/"+tt.bikeIDParam+"/images/confirm",
				strings.NewReader(tt.payload))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.bikeIDParam)
			c.Set("user", tt.userClaims)

			mockBikeRepo := new(mocks.MockBikeRepository)
			mockImageRepo := new(mocks.MockImageRepository)
			tt.mockBikeRepo(mockBikeRepo)
			tt.mockImageRepo(mockImageRepo)

			h := &ImageHandler{service: buildImageService(mockBikeRepo, mockImageRepo)}

			if assert.NoError(t, h.ConfirmUpload(c)) {
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}

			mockBikeRepo.AssertExpectations(t)
			mockImageRepo.AssertExpectations(t)
		})
	}
}
