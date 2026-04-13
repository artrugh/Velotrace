package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/velotrace/bikes-api/internal/domain"
)

var ErrBikeNotFound = errors.New("bike not found")

type PgBikeRepository struct {
	pool *pgxpool.Pool
}

func NewPgBikeRepository(pool *pgxpool.Pool) *PgBikeRepository {
	return &PgBikeRepository{pool: pool}
}

func (r *PgBikeRepository) GetAll(ctx context.Context, filter domain.BikeFilter) ([]domain.Bike, int, error) {
	var args []interface{}
	var where []string

	if filter.Status != nil {
		args = append(args, *filter.Status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	}

	if filter.CurrentOwnerID != nil {
		args = append(args, *filter.CurrentOwnerID)
		where = append(where, fmt.Sprintf("current_owner_id = $%d", len(args)))
	}

	query := "SELECT id, make_model, year, price, location_city, current_owner_id, serial_number, description, status, created_at, updated_at, COUNT(*) OVER() AS total_count FROM bikes"

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	query += " ORDER BY created_at DESC, id DESC"

	if filter.Limit > 0 {
		args = append(args, filter.Limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}

	if filter.Offset > 0 {
		args = append(args, filter.Offset)
		query += fmt.Sprintf(" OFFSET $%d", len(args))
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var bikes []domain.Bike
	var totalCount int
	for rows.Next() {
		var b domain.Bike
		if err := rows.Scan(&b.ID, &b.MakeModel, &b.Year, &b.Price, &b.LocationCity, &b.CurrentOwnerID, &b.SerialNumber, &b.Description, &b.Status, &b.CreatedAt, &b.UpdatedAt, &totalCount); err != nil {
			return nil, 0, err
		}
		bikes = append(bikes, b)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return bikes, totalCount, nil
}

func (r *PgBikeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bike, error) {
	var b domain.Bike
	err := r.pool.QueryRow(ctx, `
		SELECT id, make_model, year, price, location_city, current_owner_id, serial_number, description, status, created_at, updated_at
		FROM bikes WHERE id = $1
	`, id).Scan(&b.ID, &b.MakeModel, &b.Year, &b.Price, &b.LocationCity, &b.CurrentOwnerID, &b.SerialNumber, &b.Description, &b.Status, &b.CreatedAt, &b.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w", ErrBikeNotFound)
		}
		return nil, err
	}
	return &b, nil
}

func (r *PgBikeRepository) Create(ctx context.Context, bike *domain.Bike) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	err = tx.QueryRow(ctx, `
		INSERT INTO bikes (make_model, year, price, location_city, current_owner_id, serial_number, description, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`, bike.MakeModel, bike.Year, bike.Price, bike.LocationCity, bike.CurrentOwnerID, bike.SerialNumber, bike.Description, bike.Status).Scan(&bike.ID, &bike.CreatedAt, &bike.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("serial number already registered")
		}
		return err
	}

	_, err = tx.Exec(ctx, "INSERT INTO ownership_records (bike_id, owner_id, is_active) VALUES ($1, $2, true)", bike.ID, bike.CurrentOwnerID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PgBikeRepository) GetBikeImages(ctx context.Context, bikeID uuid.UUID) ([]domain.BikeImage, error) {
	rows, err := r.pool.Query(ctx, "SELECT id, bike_id, object_key, is_primary, created_at FROM bike_images WHERE bike_id = $1", bikeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []domain.BikeImage
	for rows.Next() {
		var img domain.BikeImage
		if err := rows.Scan(&img.ID, &img.BikeID, &img.ObjectKey, &img.IsPrimary, &img.CreatedAt); err != nil {
			return nil, err
		}
		img.PopulateURL()
		images = append(images, img)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return images, nil
}
