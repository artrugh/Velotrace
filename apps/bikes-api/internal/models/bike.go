package models

import (
	"time"

	"github.com/google/uuid"
)

type BikeStatus string

const (
	StatusRegistered  BikeStatus = "registered"
	StatusForSale     BikeStatus = "for_sale"
	StatusStolen      BikeStatus = "stolen"
	StatusTransferred BikeStatus = "transferred"
)

type Bike struct {
	ID             uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	MakeModel      string      `json:"make_model" validate:"required"` // Combined Make and Model
	Year           int         `json:"year"`
	Price          float64     `json:"price" gorm:"type:decimal(10,2)"`
	LocationCity   string      `json:"location_city"`
	CurrentOwnerID uuid.UUID   `json:"current_owner_id" gorm:"type:uuid;not null"` // Refers to users.id
	SerialNumber   string      `json:"serial_number" gorm:"uniqueIndex;not null" validate:"required"`
	Description    string      `json:"description"`
	Status         BikeStatus  `json:"status" gorm:"default:registered"`
	Images         []BikeImage `json:"images" gorm:"foreignKey:BikeID"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type BikeImage struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	BikeID    uuid.UUID `json:"bike_id" gorm:"type:uuid;not null"`
	URL       string    `json:"url" validate:"required"`
	IsPrimary bool      `json:"is_primary" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
}

type OwnershipRecord struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	BikeID     uuid.UUID  `json:"bike_id" gorm:"type:uuid;not null"`
	OwnerID    uuid.UUID  `json:"owner_id" gorm:"type:uuid;not null"` // Refers to users.id
	IsActive   bool       `json:"is_active" gorm:"default:true"`
	AcquiredAt time.Time  `json:"acquired_at" gorm:"default:now()"`
	SoldAt     *time.Time `json:"sold_at,omitempty"`
}
