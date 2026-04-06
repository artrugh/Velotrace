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
	ID             uuid.UUID   `json:"id"`
	MakeModel      string      `json:"make_model"`
	Year           int         `json:"year"`
	Price          float64     `json:"price"`
	LocationCity   string      `json:"location_city"`
	CurrentOwnerID uuid.UUID   `json:"current_owner_id"`
	SerialNumber   string      `json:"serial_number"`
	Description    string      `json:"description"`
	Status         BikeStatus  `json:"status"`
	Images         []BikeImage `json:"images"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type BikeImage struct {
	ID        uuid.UUID `json:"id"`
	BikeID    uuid.UUID `json:"bike_id"`
	ObjectKey string    `json:"-"`
	URL       string    `json:"url"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
}

func (img *BikeImage) PopulateURL() {
	baseURL := os.Getenv("STORAGE_PUBLIC_BASE_URL")
	bucket := os.Getenv("STORAGE_BUCKET")
	if img.ObjectKey != "" {
		img.URL = fmt.Sprintf("%s/%s/%s", baseURL, bucket, img.ObjectKey)
	}
}

type OwnershipRecord struct {
	ID         uuid.UUID  `json:"id"`
	BikeID     uuid.UUID  `json:"bike_id"`
	OwnerID    uuid.UUID  `json:"owner_id"`
	IsActive   bool       `json:"is_active"`
	AcquiredAt time.Time  `json:"acquired_at"`
	SoldAt     *time.Time `json:"sold_at,omitempty"`
}
