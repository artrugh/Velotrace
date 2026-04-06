package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/velotrace/bikes-api/internal/domain"
	"github.com/velotrace/bikes-api/internal/repository"
)

type BikeService interface {
	ListMarketplace(ctx context.Context) ([]domain.Bike, error)
	ListMyBikes(ctx context.Context, userID uuid.UUID) ([]domain.Bike, error)
	ListAdmin(ctx context.Context) ([]domain.Bike, error)
	GetBike(ctx context.Context, id uuid.UUID, userID string, role string) (*domain.Bike, error)
	RegisterBike(ctx context.Context, bike *domain.Bike) error
}

type bikeService struct {
	repo repository.BikeRepository
}

func NewBikeService(repo repository.BikeRepository) BikeService {
	return &bikeService{repo: repo}
}

func (s *bikeService) ListMarketplace(ctx context.Context) ([]domain.Bike, error) {
	bikes, err := s.repo.GetAll(ctx, "WHERE status = 'for_sale'", nil)
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
	bikes, err := s.repo.GetAll(ctx, "WHERE current_owner_id = $1", []interface{}{userID})
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
	bikes, err := s.repo.GetAll(ctx, "", nil)
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
	return s.repo.Create(ctx, bike)
}
