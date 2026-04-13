package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/velotrace/bikes-api/internal/domain"
	"github.com/velotrace/bikes-api/internal/testutil/mocks"
)

func newImageService(bikeRepo domain.BikeRepository, imageRepo domain.ImageRepository) *ImageService {
	return &ImageService{
		bikeRepo:  bikeRepo,
		imageRepo: imageRepo,
		storage:   nil, // not needed for ConfirmUpload tests
	}
}

func TestImageService_ConfirmUpload(t *testing.T) {
	bikeID := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()
	objectKey := "bikes/abc/123.jpg"

	tests := []struct {
		name              string
		bikeID            uuid.UUID
		userID            uuid.UUID
		objectKey         string
		setupEnv          func()
		mockBikeRepo      func(repo *mocks.MockBikeRepository)
		mockImageRepo     func(repo *mocks.MockImageRepository)
		expectError       string
		expectURLContains string
		expectIsPrimary   *bool
	}{
		{
			name:      "Success - first image becomes primary",
			bikeID:    bikeID,
			userID:    ownerID,
			objectKey: objectKey,
			setupEnv:  func() {},
			mockBikeRepo: func(repo *mocks.MockBikeRepository) {
				repo.On("GetByID", mock.Anything, bikeID).Return(&domain.Bike{
					ID: bikeID, CurrentOwnerID: ownerID,
				}, nil)
			},
			mockImageRepo: func(repo *mocks.MockImageRepository) {
				repo.On("GetImageCount", mock.Anything, bikeID).Return(0, nil)
				repo.On("CreateImage", mock.Anything, mock.MatchedBy(func(img *domain.BikeImage) bool {
					return img.IsPrimary == true && img.ObjectKey == objectKey && img.BikeID == bikeID
				})).Return(nil)
			},
			expectError: "",
		},
		{
			name:      "Success - subsequent image is not primary",
			bikeID:    bikeID,
			userID:    ownerID,
			objectKey: objectKey,
			setupEnv:  func() {},
			mockBikeRepo: func(repo *mocks.MockBikeRepository) {
				repo.On("GetByID", mock.Anything, bikeID).Return(&domain.Bike{
					ID: bikeID, CurrentOwnerID: ownerID,
				}, nil)
			},
			mockImageRepo: func(repo *mocks.MockImageRepository) {
				repo.On("GetImageCount", mock.Anything, bikeID).Return(2, nil)
				repo.On("CreateImage", mock.Anything, mock.MatchedBy(func(img *domain.BikeImage) bool {
					return img.IsPrimary == false && img.ObjectKey == objectKey
				})).Return(nil)
			},
			expectError: "",
		},
		{
			name:      "Error - bike not found",
			bikeID:    bikeID,
			userID:    ownerID,
			objectKey: objectKey,
			setupEnv:  func() {},
			mockBikeRepo: func(repo *mocks.MockBikeRepository) {
				repo.On("GetByID", mock.Anything, bikeID).Return(nil, errors.New("bike not found"))
			},
			mockImageRepo: func(repo *mocks.MockImageRepository) {},
			expectError:   "bike not found",
		},
		{
			name:      "Error - not the owner of this bike",
			bikeID:    bikeID,
			userID:    otherUserID,
			objectKey: objectKey,
			setupEnv:  func() {},
			mockBikeRepo: func(repo *mocks.MockBikeRepository) {
				repo.On("GetByID", mock.Anything, bikeID).Return(&domain.Bike{
					ID: bikeID, CurrentOwnerID: ownerID,
				}, nil)
			},
			mockImageRepo: func(repo *mocks.MockImageRepository) {},
			expectError:   "not the owner of this bike",
		},
		{
			name:      "Error - image count query fails",
			bikeID:    bikeID,
			userID:    ownerID,
			objectKey: objectKey,
			setupEnv:  func() {},
			mockBikeRepo: func(repo *mocks.MockBikeRepository) {
				repo.On("GetByID", mock.Anything, bikeID).Return(&domain.Bike{
					ID: bikeID, CurrentOwnerID: ownerID,
				}, nil)
			},
			mockImageRepo: func(repo *mocks.MockImageRepository) {
				repo.On("GetImageCount", mock.Anything, bikeID).Return(0, errors.New("count query failed"))
			},
			expectError: "count query failed",
		},
		{
			name:      "Error - create image fails",
			bikeID:    bikeID,
			userID:    ownerID,
			objectKey: objectKey,
			setupEnv:  func() {},
			mockBikeRepo: func(repo *mocks.MockBikeRepository) {
				repo.On("GetByID", mock.Anything, bikeID).Return(&domain.Bike{
					ID: bikeID, CurrentOwnerID: ownerID,
				}, nil)
			},
			mockImageRepo: func(repo *mocks.MockImageRepository) {
				repo.On("GetImageCount", mock.Anything, bikeID).Return(0, nil)
				repo.On("CreateImage", mock.Anything, mock.Anything).Return(errors.New("insert failed"))
			},
			expectError: "insert failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBikeRepo := new(mocks.MockBikeRepository)
			mockImageRepo := new(mocks.MockImageRepository)
			tt.mockBikeRepo(mockBikeRepo)
			tt.mockImageRepo(mockImageRepo)

			svc := newImageService(mockBikeRepo, mockImageRepo)
			_, err := svc.ConfirmUpload(context.Background(), tt.bikeID, tt.userID, tt.objectKey)

			if tt.expectError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError)
			} else {
				assert.NoError(t, err)
			}

			mockBikeRepo.AssertExpectations(t)
			mockImageRepo.AssertExpectations(t)
		})
	}
}

func TestImageService_ConfirmUpload_URLPopulated(t *testing.T) {
	t.Setenv("STORAGE_PUBLIC_BASE_URL", "https://cdn.example.com")
	t.Setenv("STORAGE_BUCKET", "velotrace")

	bikeID := uuid.New()
	ownerID := uuid.New()
	objectKey := "bikes/test/photo.jpg"

	mockBikeRepo := new(mocks.MockBikeRepository)
	mockBikeRepo.On("GetByID", mock.Anything, bikeID).Return(&domain.Bike{
		ID: bikeID, CurrentOwnerID: ownerID,
	}, nil)

	mockImageRepo := new(mocks.MockImageRepository)
	mockImageRepo.On("GetImageCount", mock.Anything, bikeID).Return(0, nil)
	mockImageRepo.On("CreateImage", mock.Anything, mock.Anything).Return(nil)

	svc := newImageService(mockBikeRepo, mockImageRepo)
	url, err := svc.ConfirmUpload(context.Background(), bikeID, ownerID, objectKey)

	assert.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/velotrace/bikes/test/photo.jpg", url)
}

func TestImageService_GetUploadURL_InvalidFilename(t *testing.T) {
	mockBikeRepo := new(mocks.MockBikeRepository)
	mockImageRepo := new(mocks.MockImageRepository)
	svc := newImageService(mockBikeRepo, mockImageRepo)

	_, _, err := svc.GetUploadURL(context.Background(), uuid.New(), "..")
	assert.ErrorIs(t, err, ErrInvalidFilename)

	_, _, err = svc.GetUploadURL(context.Background(), uuid.New(), "")
	assert.ErrorIs(t, err, ErrInvalidFilename)
}
