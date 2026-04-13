package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/velotrace/bikes-api/internal/domain"
	"github.com/velotrace/bikes-api/internal/platform"
)

type ImageService struct {
	imageRepo domain.ImageRepository
	bikeRepo  domain.BikeRepository
	storage   *platform.Storage
}

func NewImageService(imageRepo domain.ImageRepository, bikeRepo domain.BikeRepository, storage *platform.Storage) *ImageService {
	return &ImageService{
		imageRepo: imageRepo,
		bikeRepo:  bikeRepo,
		storage:   storage,
	}
}

func (s *ImageService) GetUploadURL(ctx context.Context, bikeID uuid.UUID, filename string) (string, string, error) {
	// Sanitize filename: extract base name and remove path traversal
	filename = filepath.Base(filename)
	filename = strings.ReplaceAll(filename, "..", "")
	if filename == "" || filename == "." {
		return "", "", fmt.Errorf("invalid filename")
	}

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
		if strings.Contains(err.Error(), "not found") {
			return "", ErrBikeNotFound
		}
		return "", err
	}
	if bike == nil {
		return "", ErrBikeNotFound
	}
	if bike.CurrentOwnerID != userID {
		return "", ErrNotOwner
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
