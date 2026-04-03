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
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "transaction failed"})
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
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return c.JSON(http.StatusConflict, map[string]string{"error": "serial number already registered"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create bike"})
	}

	_, err = tx.Exec(context.Background(), "INSERT INTO ownership_records (bike_id, owner_id, is_active) VALUES ($1, $2, true)", bike.ID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create ownership record"})
	}

	if err := tx.Commit(context.Background()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "commit failed"})
	}

	return c.JSON(http.StatusCreated, bike)
}

// ListMarketplace returns only bikes with status 'for_sale' (Public)
// @Summary List public bikes for sale
// @Description Returns a list of bikes with status 'for_sale'. Sensitive fields are redacted.
// @Tags bikes
// @Produce json
// @Success 200 {array} models.Bike
// @Failure 500 {object} map[string]string
// @Router /bikes [get]
func (h *BikeHandler) ListMarketplace(c echo.Context) error {
	return h.listBikes(c, "WHERE status = 'for_sale'", nil, true)
}

// ListMyBikes returns all bikes owned by the current user (Protected)
// @Summary List my bikes
// @Description Returns all bikes owned by the authenticated user with full metadata.
// @Tags bikes
// @Produce json
// @Security Bearer
// @Success 200 {array} models.Bike
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /my/bikes [get]
func (h *BikeHandler) ListMyBikes(c echo.Context) error {
	userClaims := c.Get("user").(*auth.UserClaims)
	userID, _ := uuid.Parse(userClaims.UserID)
	return h.listBikes(c, "WHERE current_owner_id = $1", []interface{}{userID}, false)
}

// ListAdmin returns every bike in the system (Admin Only)
// @Summary List all bikes (Admin)
// @Description Returns every bike in the system. Strictly for users with admin role.
// @Tags admin
// @Produce json
// @Security Bearer
// @Success 200 {array} models.Bike
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/bikes [get]
func (h *BikeHandler) ListAdmin(c echo.Context) error {
	userClaims := c.Get("user").(*auth.UserClaims)
	if userClaims.Role != "admin" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "admin access required"})
	}
	return h.listBikes(c, "", nil, false)
}

// GetBike returns a single bike with smart visibility (Public/Owner/Admin)
// @Summary Get bike by ID
// @Description Returns bike details. Hybrid logic: Public fields if for_sale, full fields if owner/admin, 404 otherwise.
// @Tags bikes
// @Produce json
// @Security Bearer
// @Param id path string true "Bike ID"
// @Success 200 {object} models.Bike
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /bikes/{id} [get]
func (h *BikeHandler) GetBike(c echo.Context) error {
	id := c.Param("id")
	userClaims := c.Get("user").(*auth.UserClaims)

	var b models.Bike
	err := h.DB.QueryRow(context.Background(), `
		SELECT id, make_model, year, price, location_city, current_owner_id, serial_number, description, status, created_at, updated_at
		FROM bikes WHERE id = $1
	`, id).Scan(&b.ID, &b.MakeModel, &b.Year, &b.Price, &b.LocationCity, &b.CurrentOwnerID, &b.SerialNumber, &b.Description, &b.Status, &b.CreatedAt, &b.UpdatedAt)

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "bike not found"})
	}

	isOwnerOrAdmin := userClaims.UserID == b.CurrentOwnerID.String() || userClaims.Role == "admin"

	// Logic Check (Silent Sentry)
	if !isOwnerOrAdmin && b.Status != models.StatusForSale {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "bike not found"})
	}

	// Sanitization Check
	if !isOwnerOrAdmin {
		b.SerialNumber = "REDACTED"
		b.CurrentOwnerID = uuid.Nil
	}

	b.Images, _ = h.getBikeImages(b.ID)
	return c.JSON(http.StatusOK, b)
}

func (h *BikeHandler) listBikes(c echo.Context, whereClause string, args []interface{}, sanitize bool) error {
	query := "SELECT id, make_model, year, price, location_city, current_owner_id, serial_number, description, status, created_at, updated_at FROM bikes " + whereClause

	rows, err := h.DB.Query(context.Background(), query, args...)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "fetch failed"})
	}
	defer rows.Close()

	var bikes []models.Bike
	for rows.Next() {
		var b models.Bike
		if err := rows.Scan(&b.ID, &b.MakeModel, &b.Year, &b.Price, &b.LocationCity, &b.CurrentOwnerID, &b.SerialNumber, &b.Description, &b.Status, &b.CreatedAt, &b.UpdatedAt); err == nil {
			if sanitize {
				b.SerialNumber = "REDACTED"
				b.CurrentOwnerID = uuid.Nil
			}
			b.Images, _ = h.getBikeImages(b.ID)
			bikes = append(bikes, b)
		}
	}
	return c.JSON(http.StatusOK, bikes)
}

func (h *BikeHandler) getBikeImages(bikeID uuid.UUID) ([]models.BikeImage, error) {
	rows, err := h.DB.Query(context.Background(), "SELECT id, bike_id, object_key, is_primary, created_at FROM bike_images WHERE bike_id = $1", bikeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []models.BikeImage
	for rows.Next() {
		var img models.BikeImage
		if err := rows.Scan(&img.ID, &img.BikeID, &img.ObjectKey, &img.IsPrimary, &img.CreatedAt); err == nil {
			img.PopulateURL()
			images = append(images, img)
		}
	}
	return images, nil
}
