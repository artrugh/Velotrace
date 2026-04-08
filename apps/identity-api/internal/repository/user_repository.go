package repository

import (
	"context"

	"github.com/velotrace/identity-api/internal/domain"
)

type UserRepository interface {
	UpsertByGoogleID(ctx context.Context, googleID, email, displayName string) (*domain.User, error)
}
