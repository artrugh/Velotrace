package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/velotrace/bikes-api/internal/domain"
)

var (
	ErrSerialNumberExists = errors.New("serial number already registered")
)

type BikeFilter struct {
	Status         *domain.BikeStatus
	CurrentOwnerID *uuid.UUID
}

type BikeRepository interface {
	GetAll(ctx context.Context, filter BikeFilter) ([]domain.Bike, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Bike, error)
	Create(ctx context.Context, bike *domain.Bike) error
	GetBikeImages(ctx context.Context, bikeID uuid.UUID) ([]domain.BikeImage, error)
}

type BikeService interface {
	ListMarketplace(ctx context.Context) ([]domain.Bike, error)
	ListMyBikes(ctx context.Context, userID uuid.UUID) ([]domain.Bike, error)
	ListAdmin(ctx context.Context) ([]domain.Bike, error)
	GetBike(ctx context.Context, id uuid.UUID, userID string, role string) (*domain.Bike, error)
	RegisterBike(ctx context.Context, bike *domain.Bike) error
}

type bikeService struct {
	repo BikeRepository
}

func NewBikeService(repo BikeRepository) BikeService {
	return &bikeService{repo: repo}
}

func (s *bikeService) ListMarketplace(ctx context.Context) ([]domain.Bike, error) {
	status := domain.StatusForSale
	bikes, err := s.repo.GetAll(ctx, BikeFilter{Status: &status})
	if err != nil {
		return nil, err
	}

	for i := range bikes {
		bikes[i].SerialNumber = "REDACTED"
		bikes[i].CurrentOwnerID = uuid.Nil
		images, _ := s.repo.GetBikeImages(ctx, bikes[i].ID)
		bikes[i].Images = images
	}

	return bikes, nil
}

func (s *bikeService) ListMyBikes(ctx context.Context, userID uuid.UUID) ([]domain.Bike, error) {
	bikes, err := s.repo.GetAll(ctx, BikeFilter{CurrentOwnerID: &userID})
	if err != nil {
		return nil, err
	}

	for i := range bikes {
		images, _ := s.repo.GetBikeImages(ctx, bikes[i].ID)
		bikes[i].Images = images
	}

	return bikes, nil
}

func (s *bikeService) ListAdmin(ctx context.Context) ([]domain.Bike, error) {
	bikes, err := s.repo.GetAll(ctx, BikeFilter{})
	if err != nil {
		return nil, err
	}

	for i := range bikes {
		images, _ := s.repo.GetBikeImages(ctx, bikes[i].ID)
		bikes[i].Images = images
	}

	return bikes, nil
}

func (s *bikeService) GetBike(ctx context.Context, id uuid.UUID, userID string, role string) (*domain.Bike, error) {
	bike, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if bike == nil {
		return nil, fmt.Errorf("bike not found")
	}

	isOwnerOrAdmin := userID == bike.CurrentOwnerID.String() || role == "admin"

	if !isOwnerOrAdmin && bike.Status != domain.StatusForSale {
		return nil, fmt.Errorf("bike not found")
	}

	if !isOwnerOrAdmin {
		bike.SerialNumber = "REDACTED"
		bike.CurrentOwnerID = uuid.Nil
	}

	images, _ := s.repo.GetBikeImages(ctx, bike.ID)
	bike.Images = images

	return bike, nil
}

func (s *bikeService) RegisterBike(ctx context.Context, bike *domain.Bike) error {
	bike.Status = domain.StatusRegistered
	err := s.repo.Create(ctx, bike)
	if err != nil && err.Error() == "serial number already registered" {
		return fmt.Errorf("%w", ErrSerialNumberExists)
	}
	return err
}
