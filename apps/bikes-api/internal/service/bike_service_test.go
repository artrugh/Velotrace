package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/velotrace/bikes-api/internal/domain"
	"github.com/velotrace/bikes-api/internal/repository"
)

func TestBikeService_GetBike(t *testing.T) {
	bikeID := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New().String()

	tests := []struct {
		name          string
		id            uuid.UUID
		userID        string
		role          string
		mockBehavior  func(repo *repository.MockBikeRepository)
		expectedBike  *domain.Bike
		expectedError string
	}{
		{
			name:   "Success - Owner Access",
			id:     bikeID,
			userID: ownerID.String(),
			role:   "user",
			mockBehavior: func(repo *repository.MockBikeRepository) {
				repo.On("GetByID", mock.Anything, bikeID).Return(&domain.Bike{ID: bikeID, CurrentOwnerID: ownerID, Status: domain.StatusRegistered}, nil)
				repo.On("GetBikeImages", mock.Anything, bikeID).Return([]domain.BikeImage{}, nil)
			},
			expectedBike: &domain.Bike{ID: bikeID, CurrentOwnerID: ownerID, Status: domain.StatusRegistered, Images: []domain.BikeImage{}},
		},
		{
			name:   "Success - Public Access (Redacted)",
			id:     bikeID,
			userID: otherUserID,
			role:   "user",
			mockBehavior: func(repo *repository.MockBikeRepository) {
				repo.On("GetByID", mock.Anything, bikeID).Return(&domain.Bike{ID: bikeID, CurrentOwnerID: ownerID, Status: domain.StatusForSale, SerialNumber: "SN123"}, nil)
				repo.On("GetBikeImages", mock.Anything, bikeID).Return([]domain.BikeImage{}, nil)
			},
			expectedBike: &domain.Bike{ID: bikeID, CurrentOwnerID: uuid.Nil, Status: domain.StatusForSale, SerialNumber: "REDACTED", Images: []domain.BikeImage{}},
		},
		{
			name:   "Error - Not Found (Private Bike)",
			id:     bikeID,
			userID: otherUserID,
			role:   "user",
			mockBehavior: func(repo *repository.MockBikeRepository) {
				repo.On("GetByID", mock.Anything, bikeID).Return(&domain.Bike{ID: bikeID, CurrentOwnerID: ownerID, Status: domain.StatusRegistered}, nil)
			},
			expectedError: "bike not found",
		},
		{
			name:   "Error - DB Failure",
			id:     bikeID,
			userID: ownerID.String(),
			role:   "user",
			mockBehavior: func(repo *repository.MockBikeRepository) {
				repo.On("GetByID", mock.Anything, bikeID).Return(nil, errors.New("db connection failed"))
			},
			expectedError: "db connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repository.MockBikeRepository)
			tt.mockBehavior(mockRepo)
			svc := NewBikeService(mockRepo)

			result, err := svc.GetBike(context.Background(), tt.id, tt.userID, tt.role)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBike, result)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBikeService_RegisterBike(t *testing.T) {
	tests := []struct {
		name         string
		bike         *domain.Bike
		mockBehavior func(repo *repository.MockBikeRepository)
		expectedErr  error
	}{
		{
			name: "Success",
			bike: &domain.Bike{MakeModel: "Trek Domane", SerialNumber: "123"},
			mockBehavior: func(repo *repository.MockBikeRepository) {
				repo.On("Create", mock.Anything, &domain.Bike{MakeModel: "Trek Domane", SerialNumber: "123", Status: domain.StatusRegistered}).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "Error - Duplicate Serial",
			bike: &domain.Bike{MakeModel: "Trek Domane", SerialNumber: "123"},
			mockBehavior: func(repo *repository.MockBikeRepository) {
				repo.On("Create", mock.Anything, &domain.Bike{MakeModel: "Trek Domane", SerialNumber: "123", Status: domain.StatusRegistered}).Return(errors.New("serial number already registered"))
			},
			expectedErr: ErrSerialNumberExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repository.MockBikeRepository)
			tt.mockBehavior(mockRepo)
			svc := NewBikeService(mockRepo)

			err := svc.RegisterBike(context.Background(), tt.bike)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, domain.StatusRegistered, tt.bike.Status)
			mockRepo.AssertExpectations(t)
		})
	}
}
