package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/velotrace/bikes-api/internal/domain"
)

type PgImageRepository struct {
	pool *pgxpool.Pool
}

func NewPgImageRepository(pool *pgxpool.Pool) *PgImageRepository {
	return &PgImageRepository{pool: pool}
}

func (r *PgImageRepository) GetImageCount(ctx context.Context, bikeID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM bike_images WHERE bike_id = $1", bikeID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("repo GetImageCount: %w", err)
	}
	return count, nil
}

func (r *PgImageRepository) CreateImage(ctx context.Context, img *domain.BikeImage) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO bike_images (bike_id, object_key, is_primary)
		VALUES ($1, $2, $3)
	`, img.BikeID, img.ObjectKey, img.IsPrimary)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return domain.ErrBikeNotFound
		}

		return fmt.Errorf("repo CreateImage: %w", err)
	}
	return nil
}
