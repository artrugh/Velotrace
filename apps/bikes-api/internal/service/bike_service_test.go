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
		name           string
		id             uuid.UUID
		userID         string
		role           string
		mockBehavior   func(repo *repository.MockBikeRepository)
		expectedBike   *domain.Bike
		expectedError  string
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
	bike := &domain.Bike{MakeModel: "Trek Domane", SerialNumber: "123"}

	tests := []struct {
		name         string
		bike         *domain.Bike
		mockBehavior func(repo *repository.MockBikeRepository)
		expectedErr  error
	}{
		{
			name: "Success",
			bike: bike,
			mockBehavior: func(repo *repository.MockBikeRepository) {
				repo.On("Create", mock.Anything, bike).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "Error - Duplicate Serial",
			bike: bike,
			mockBehavior: func(repo *repository.MockBikeRepository) {
				repo.On("Create", mock.Anything, bike).Return(errors.New("serial number already registered"))
			},
			expectedErr: errors.New("serial number already registered"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repository.MockBikeRepository)
			tt.mockBehavior(mockRepo)
			svc := NewBikeService(mockRepo)

			err := svc.RegisterBike(context.Background(), tt.bike)

			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, domain.StatusRegistered, tt.bike.Status)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBikeService_ListMarketplace(t *testing.T) {
	bikeID1 := uuid.New()
	bikeID2 := uuid.New()

	tests := []struct {
		name          string
		mockBehavior  func(repo *repository.MockBikeRepository)
		expectedCount int
		expectRedact  bool
		expectError   bool
	}{
		{
			name: "Returns bikes with sensitive fields redacted",
			mockBehavior: func(repo *repository.MockBikeRepository) {
				repo.On("GetAll", mock.Anything, "WHERE status = 'for_sale'", []interface{}(nil)).
					Return([]domain.Bike{
						{ID: bikeID1, SerialNumber: "SN-001", CurrentOwnerID: uuid.New(), Status: domain.StatusForSale},
						{ID: bikeID2, SerialNumber: "SN-002", CurrentOwnerID: uuid.New(), Status: domain.StatusForSale},
					}, nil)
				repo.On("GetBikeImages", mock.Anything, bikeID1).Return([]domain.BikeImage{}, nil)
				repo.On("GetBikeImages", mock.Anything, bikeID2).Return([]domain.BikeImage{}, nil)
			},
			expectedCount: 2,
			expectRedact:  true,
		},
		{
			name: "Returns empty list when no bikes for sale",
			mockBehavior: func(repo *repository.MockBikeRepository) {
				repo.On("GetAll", mock.Anything, "WHERE status = 'for_sale'", []interface{}(nil)).
					Return([]domain.Bike{}, nil)
			},
			expectedCount: 0,
		},
		{
			name: "Returns error on repository failure",
			mockBehavior: func(repo *repository.MockBikeRepository) {
				repo.On("GetAll", mock.Anything, "WHERE status = 'for_sale'", []interface{}(nil)).
					Return(nil, errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repository.MockBikeRepository)
			tt.mockBehavior(mockRepo)
			svc := NewBikeService(mockRepo)

			bikes, err := svc.ListMarketplace(context.Background())

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, bikes, tt.expectedCount)
				if tt.expectRedact {
					for _, b := range bikes {
						assert.Equal(t, "REDACTED", b.SerialNumber)
						assert.Equal(t, uuid.Nil, b.CurrentOwnerID)
					}
				}
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBikeService_ListMyBikes(t *testing.T) {
	userID := uuid.New()
	bikeID := uuid.New()

	tests := []struct {
		name          string
		userID        uuid.UUID
		mockBehavior  func(repo *repository.MockBikeRepository)
		expectedCount int
		expectError   bool
	}{
		{
			name:   "Returns bikes owned by user with full data",
			userID: userID,
			mockBehavior: func(repo *repository.MockBikeRepository) {
				repo.On("GetAll", mock.Anything, "WHERE current_owner_id = $1", []interface{}{userID}).
					Return([]domain.Bike{
						{ID: bikeID, SerialNumber: "SN-REAL", CurrentOwnerID: userID, Status: domain.StatusRegistered},
					}, nil)
				repo.On("GetBikeImages", mock.Anything, bikeID).Return([]domain.BikeImage{}, nil)
			},
			expectedCount: 1,
		},
		{
			name:   "Serial number is not redacted for owner",
			userID: userID,
			mockBehavior: func(repo *repository.MockBikeRepository) {
				repo.On("GetAll", mock.Anything, "WHERE current_owner_id = $1", []interface{}{userID}).
					Return([]domain.Bike{
						{ID: bikeID, SerialNumber: "SECRET-SN", CurrentOwnerID: userID},
					}, nil)
				repo.On("GetBikeImages", mock.Anything, bikeID).Return(nil, nil)
			},
			expectedCount: 1,
		},
		{
			name:   "Returns error on repository failure",
			userID: userID,
			mockBehavior: func(repo *repository.MockBikeRepository) {
				repo.On("GetAll", mock.Anything, "WHERE current_owner_id = $1", []interface{}{userID}).
					Return(nil, errors.New("connection failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repository.MockBikeRepository)
			tt.mockBehavior(mockRepo)
			svc := NewBikeService(mockRepo)

			bikes, err := svc.ListMyBikes(context.Background(), tt.userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, bikes, tt.expectedCount)
				if len(bikes) > 0 {
					assert.NotEqual(t, "REDACTED", bikes[0].SerialNumber)
				}
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBikeService_ListAdmin(t *testing.T) {
	bikeID1 := uuid.New()
	bikeID2 := uuid.New()

	tests := []struct {
		name          string
		mockBehavior  func(repo *repository.MockBikeRepository)
		expectedCount int
		expectError   bool
	}{
		{
			name: "Returns all bikes with no filter",
			mockBehavior: func(repo *repository.MockBikeRepository) {
				repo.On("GetAll", mock.Anything, "", []interface{}(nil)).
					Return([]domain.Bike{
						{ID: bikeID1, Status: domain.StatusRegistered},
						{ID: bikeID2, Status: domain.StatusForSale},
					}, nil)
				repo.On("GetBikeImages", mock.Anything, bikeID1).Return([]domain.BikeImage{}, nil)
				repo.On("GetBikeImages", mock.Anything, bikeID2).Return([]domain.BikeImage{}, nil)
			},
			expectedCount: 2,
		},
		{
			name: "Returns error on repository failure",
			mockBehavior: func(repo *repository.MockBikeRepository) {
				repo.On("GetAll", mock.Anything, "", []interface{}(nil)).
					Return(nil, errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repository.MockBikeRepository)
			tt.mockBehavior(mockRepo)
			svc := NewBikeService(mockRepo)

			bikes, err := svc.ListAdmin(context.Background())

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, bikes, tt.expectedCount)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBikeService_GetBike_AdminAccess(t *testing.T) {
	bikeID := uuid.New()
	ownerID := uuid.New()
	adminID := uuid.New().String()

	mockRepo := new(repository.MockBikeRepository)
	mockRepo.On("GetByID", mock.Anything, bikeID).Return(&domain.Bike{
		ID: bikeID, CurrentOwnerID: ownerID, Status: domain.StatusRegistered, SerialNumber: "ADMIN-SN",
	}, nil)
	mockRepo.On("GetBikeImages", mock.Anything, bikeID).Return([]domain.BikeImage{}, nil)

	svc := NewBikeService(mockRepo)
	result, err := svc.GetBike(context.Background(), bikeID, adminID, "admin")

	assert.NoError(t, err)
	// Admin sees full data - serial number not redacted
	assert.Equal(t, "ADMIN-SN", result.SerialNumber)
	assert.Equal(t, ownerID, result.CurrentOwnerID)
	mockRepo.AssertExpectations(t)
}