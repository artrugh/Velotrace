package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/velotrace/bikes-api/internal/domain"
	"github.com/velotrace/bikes-api/internal/service"
	"velotrace.local/auth"
)

type BikeHandler struct {
	service *service.BikeService
}

func NewBikeHandler(service *service.BikeService) *BikeHandler {
	return &BikeHandler{service: service}
}

type RegisterBikeRequest struct {
	MakeModel    string  `json:"make_model"`
	Year         int     `json:"year"`
	Price        float64 `json:"price"`
	LocationCity string  `json:"location_city"`
	SerialNumber string  `json:"serial_number"`
	Description  string  `json:"description"`
}

// RegisterBike registers a new bike and sets the current user as the owner
// @Summary Register a new bike
// @Description Creates a bike entry and an ownership record in a single transaction
// @Tags bikes
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body RegisterBikeRequest true "Bike registration data"
// @Success 201 {object} domain.Bike
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

	bike := &domain.Bike{
		MakeModel:      req.MakeModel,
		Year:           req.Year,
		Price:          req.Price,
		LocationCity:   req.LocationCity,
		SerialNumber:   req.SerialNumber,
		Description:    req.Description,
		CurrentOwnerID: userID,
	}

	if err := h.service.RegisterBike(c.Request().Context(), bike); err != nil {
		if err.Error() == "serial number already registered" {
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create bike"})
	}

	return c.JSON(http.StatusCreated, bike)
}

// ListMarketplace returns only bikes with status 'for_sale' (Public)
// @Summary List public bikes for sale
// @Description Returns a list of bikes with status 'for_sale'. Sensitive fields are redacted.
// @Tags bikes
// @Produce json
// @Success 200 {array} domain.Bike
// @Failure 500 {object} map[string]string
// @Router /bikes [get]
func (h *BikeHandler) ListMarketplace(c echo.Context) error {
	bikes, err := h.service.ListMarketplace(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "fetch failed"})
	}
	return c.JSON(http.StatusOK, bikes)
}

// ListMyBikes returns all bikes owned by the current user (Protected)
// @Summary List my bikes
// @Description Returns all bikes owned by the authenticated user with full metadata.
// @Tags bikes
// @Produce json
// @Security Bearer
// @Success 200 {array} domain.Bike
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /my/bikes [get]
func (h *BikeHandler) ListMyBikes(c echo.Context) error {
	userClaims := c.Get("user").(*auth.UserClaims)
	userID, _ := uuid.Parse(userClaims.UserID)
	bikes, err := h.service.ListMyBikes(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "fetch failed"})
	}
	return c.JSON(http.StatusOK, bikes)
}

// ListAdmin returns every bike in the system (Admin Only)
// @Summary List all bikes (Admin)
// @Description Returns every bike in the system. Strictly for users with admin role.
// @Tags admin
// @Produce json
// @Security Bearer
// @Success 200 {array} domain.Bike
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/bikes [get]
func (h *BikeHandler) ListAdmin(c echo.Context) error {
	userClaims := c.Get("user").(*auth.UserClaims)
	if userClaims.Role != "admin" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "admin access required"})
	}
	bikes, err := h.service.ListAdmin(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "fetch failed"})
	}
	return c.JSON(http.StatusOK, bikes)
}

// GetBike returns a single bike with smart visibility (Public/Owner/Admin)
// @Summary Get bike by ID
// @Description Returns bike details. Hybrid logic: Public fields if for_sale, full fields if owner/admin, 404 otherwise.
// @Tags bikes
// @Produce json
// @Security Bearer
// @Param id path string true "Bike ID"
// @Success 200 {object} domain.Bike
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /bikes/{id} [get]
func (h *BikeHandler) GetBike(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid bike id"})
	}

	userClaims := c.Get("user").(*auth.UserClaims)

	bike, err := h.service.GetBike(c.Request().Context(), id, userClaims.UserID, userClaims.Role)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "bike not found"})
	}

	return c.JSON(http.StatusOK, bike)
}
