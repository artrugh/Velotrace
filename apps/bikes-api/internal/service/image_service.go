package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/velotrace/bikes-api/internal/domain"
	"github.com/velotrace/bikes-api/internal/platform"
	"github.com/velotrace/bikes-api/internal/repository"
)

type ImageService struct {
	imageRepo repository.ImageRepository
	bikeRepo  repository.BikeRepository
	storage   *platform.Storage
}

func NewImageService(imageRepo repository.ImageRepository, bikeRepo repository.BikeRepository, storage *platform.Storage) *ImageService {
	return &ImageService{
		imageRepo: imageRepo,
		bikeRepo:  bikeRepo,
		storage:   storage,
	}
}

func (s *ImageService) GetUploadURL(ctx context.Context, bikeID uuid.UUID, filename string) (string, string, error) {
	timestamp := time.Now().Unix()
	objectKey := fmt.Sprintf("bikes/%s/%d_%s", bikeID, timestamp, filename)

	uploadURL, err := s.storage.GetPresignedPutURL(ctx, objectKey)
	if err != nil {
		return "", "", err
	}

	return uploadURL, objectKey, nil
}

func (s *ImageService) ConfirmUpload(ctx context.Context, bikeID uuid.UUID, userID uuid.UUID, objectKey string) (string, error) {
	// Verify ownership
	bike, err := s.bikeRepo.GetByID(ctx, bikeID)
	if err != nil {
		return "", fmt.Errorf("bike not found")
	}
	if bike.CurrentOwnerID != userID {
		return "", fmt.Errorf("not the owner of this bike")
	}

	// Check image count for primary status
	count, err := s.imageRepo.GetImageCount(ctx, bikeID)
	if err != nil {
		return "", err
	}

	isPrimary := count == 0
	img := &domain.BikeImage{
		BikeID:    bikeID,
		ObjectKey: objectKey,
		IsPrimary: isPrimary,
	}

	if err := s.imageRepo.CreateImage(ctx, img); err != nil {
		return "", err
	}

	img.PopulateURL()
	return img.URL, nil
}
