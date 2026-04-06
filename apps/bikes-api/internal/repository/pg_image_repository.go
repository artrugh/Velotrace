package repository

import (
	"context"

	"github.com/google/uuid"
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
	return count, err
}

func (r *PgImageRepository) CreateImage(ctx context.Context, img *domain.BikeImage) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO bike_images (bike_id, object_key, is_primary)
		VALUES ($1, $2, $3)
	`, img.BikeID, img.ObjectKey, img.IsPrimary)
	return err
}
