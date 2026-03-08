package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	GoogleID    string    `json:"-"`
	DisplayName string    `json:"display_name"`
	FirstName   *string   `json:"first_name"`
	LastName    *string   `json:"last_name"`
	IsVerified  bool      `json:"is_verified"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
