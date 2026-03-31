package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/velotrace/bikes-api/internal/models"
	"velotrace.local/auth"
)

type BikeHandler struct {
	DB *pgxpool.Pool
}

type RegisterBikeRequest struct {
	MakeModel    string   `json:"make_model"`
	Year         int      `json:"year"`
	Price        float64  `json:"price"`
	LocationCity string   `json:"location_city"`
	SerialNumber string   `json:"serial_number"`
	Description  string   `json:"description"`
	ImageKeys    []string `json:"image_keys"`
}

// RegisterBike registers a new bike and sets the current user as the owner
// @Summary Register a new bike
// @Description Creates a bike entry and an ownership record in a single transaction
// @Tags bikes
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body RegisterBikeRequest true "Bike registration data"
// @Success 201 {object} models.Bike
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /bikes [post]
func (h *BikeHandler) RegisterBike(c echo.Context) error {
	userClaims := c.Get("user").(*auth.UserClaims)
	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user ID"})
	}

	var req RegisterBikeRequest
	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	tx, err := h.DB.Begin(context.Background())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to start transaction"})
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	var bike models.Bike
	err = tx.QueryRow(context.Background(), `
		INSERT INTO bikes (make_model, year, price, location_city, current_owner_id, serial_number, description, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'registered')
		RETURNING id, make_model, year, price, location_city, current_owner_id, serial_number, description, status, created_at, updated_at
	`, req.MakeModel, req.Year, req.Price, req.LocationCity, userID, req.SerialNumber, req.Description).Scan(
		&bike.ID, &bike.MakeModel, &bike.Year, &bike.Price, &bike.LocationCity, &bike.CurrentOwnerID, &bike.SerialNumber, &bike.Description, &bike.Status, &bike.CreatedAt, &bike.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		// Check if it's a Postgres-specific error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				// Return 409 Conflict instead of 500
				return c.JSON(http.StatusConflict, map[string]string{
					"error": "A bike with this serial number is already registered",
				})
			}
		}

		// If it's not a duplicate key error, return the 500 you had before
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "failed to create bike",
			"details": err.Error(),
		})
	}

	// Create ownership record
	_, err = tx.Exec(context.Background(), `
		INSERT INTO ownership_records (bike_id, owner_id, is_active)
		VALUES ($1, $2, true)
	`, bike.ID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create ownership record"})
	}

	// Save image keys
	for i, key := range req.ImageKeys {
		_, err = tx.Exec(context.Background(), `
			INSERT INTO bike_images (bike_id, object_key, is_primary)
			VALUES ($1, $2, $3)
		`, bike.ID, key, i == 0)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save bike images"})
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to commit transaction"})
	}

	return c.JSON(http.StatusCreated, bike)
}

// ListBikesPublic returns all bikes currently for sale
// @Summary List public bikes for sale
// @Description Returns a list of bikes with status 'for_sale' including images
// @Tags bikes
// @Produce json
// @Success 200 {array} models.Bike
// @Failure 500 {object} map[string]string
// @Router /bikes [get]
func (h *BikeHandler) ListBikesPublic(c echo.Context) error {
	rows, err := h.DB.Query(context.Background(), `
		SELECT id, make_model, year, price, location_city, current_owner_id, serial_number, description, status, created_at, updated_at
		FROM bikes
		WHERE status = 'for_sale'
	`)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch bikes"})
	}
	defer rows.Close()

	var bikes []models.Bike
	for rows.Next() {
		var b models.Bike
		if err := rows.Scan(&b.ID, &b.MakeModel, &b.Year, &b.Price, &b.LocationCity, &b.CurrentOwnerID, &b.SerialNumber, &b.Description, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to scan bike"})
		}

		// Preload images for each bike
		imgRows, err := h.DB.Query(context.Background(), "SELECT id, bike_id, object_key, is_primary, created_at FROM bike_images WHERE bike_id = $1", b.ID)
		if err == nil {
			var images []models.BikeImage
			for imgRows.Next() {
				var img models.BikeImage
				if err := imgRows.Scan(&img.ID, &img.BikeID, &img.ObjectKey, &img.IsPrimary, &img.CreatedAt); err == nil {
					img.PopulateURL()
					images = append(images, img)
				}
			}
			imgRows.Close()
			b.Images = images
		}

		bikes = append(bikes, b)
	}

	return c.JSON(http.StatusOK, bikes)
}
