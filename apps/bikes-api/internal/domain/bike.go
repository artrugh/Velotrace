package domain

import (
	"fmt"
	"os"
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
	ID             uuid.UUID   `json:"id"`                                // Primary Key (UUID)
	MakeModel      string      `json:"make_model" validate:"required"`    // Combined Make and Model (Required)
	Year           int         `json:"year"`                              // Manufacturing Year
	Price          float64     `json:"price"`                             // Decimal price (Mapped to float64)
	LocationCity   string      `json:"location_city"`                     // City where the bike is located
	CurrentOwnerID uuid.UUID   `json:"current_owner_id"`                  // FK to Users (Not Null)
	SerialNumber   string      `json:"serial_number" validate:"required"` // Unique Serial Number (Required)
	Description    string      `json:"description"`                       // Optional text description
	Status         BikeStatus  `json:"status"`                            // Enum (registered, for_sale, stolen, transferred)
	Images         []BikeImage `json:"images" validate:"max=20" maxItems:"20"` // Related images (Max 20)
	CreatedAt      time.Time   `json:"created_at"`                        // Timestamp
	UpdatedAt      time.Time   `json:"updated_at"`                        // Timestamp
}

type BikeImage struct {
	ID        uuid.UUID `json:"id"`         // Primary Key
	BikeID    uuid.UUID `json:"bike_id"`    // FK to Bikes
	ObjectKey string    `json:"-"`          // S3/Minio object key (Hidden from JSON)
	URL       string    `json:"url"`        // Publicly accessible URL (Calculated)
	IsPrimary bool      `json:"is_primary"` // Boolean flag (Default: false)
	CreatedAt time.Time `json:"created_at"` // Timestamp
}

func (img *BikeImage) PopulateURL() {
	baseURL := os.Getenv("STORAGE_PUBLIC_BASE_URL")
	bucket := os.Getenv("STORAGE_BUCKET")
	if img.ObjectKey != "" {
		img.URL = fmt.Sprintf("%s/%s/%s", baseURL, bucket, img.ObjectKey)
	}
}

type OwnershipRecord struct {
	ID         uuid.UUID  `json:"id"`                // Primary Key
	BikeID     uuid.UUID  `json:"bike_id"`           // FK to Bikes
	OwnerID    uuid.UUID  `json:"owner_id"`          // FK to Users
	IsActive   bool       `json:"is_active"`         // Boolean flag (Default: true)
	AcquiredAt time.Time  `json:"acquired_at"`       // Timestamp (Default: now())
	SoldAt     *time.Time `json:"sold_at,omitempty"` // Optional Timestamp
}
