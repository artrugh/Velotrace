package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email       string    `json:"email" example:"user@example.com"`
	GoogleID    string    `json:"-"`
	DisplayName string    `json:"display_name" example:"John Doe"`
	FirstName   *string   `json:"first_name,omitempty" example:"John"`
	LastName    *string   `json:"last_name,omitempty" example:"Doe"`
	IsVerified  bool      `json:"is_verified" example:"true"`
	LastLogin   time.Time `json:"last_login"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
