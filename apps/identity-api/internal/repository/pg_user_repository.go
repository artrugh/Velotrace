package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/velotrace/identity-api/internal/domain"
	"github.com/velotrace/identity-api/internal/service"
)

type pgUserRepository struct {
	pool *pgxpool.Pool
}

func NewPgUserRepository(pool *pgxpool.Pool) service.UserRepository {
	return &pgUserRepository{pool: pool}
}

func (r *pgUserRepository) UpsertByGoogleID(ctx context.Context, googleID, email, displayName string) (*domain.User, error) {
	var user domain.User
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (google_id, email, display_name, last_login, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (google_id) DO UPDATE SET
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			last_login = NOW(),
			updated_at = NOW()
		RETURNING id, email, display_name, first_name, last_name, is_verified, role, last_login, created_at, updated_at, google_id
	`, googleID, email, displayName).Scan(
		&user.ID, &user.Email, &user.DisplayName, &user.FirstName, &user.LastName, &user.IsVerified, &user.Role, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt, &user.GoogleID,
	)

	if err != nil {
		return nil, err
	}
	return &user, nil
}
