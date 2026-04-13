package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/velotrace/bikes-api/internal/domain"
	"github.com/velotrace/bikes-api/internal/service"
	"velotrace.local/auth"
)

type BikeHandler struct {
	service service.BikeService
}

func NewBikeHandler(service service.BikeService) *BikeHandler {
	return &BikeHandler{service: service}
}

type RegisterBikeRequest struct {
	MakeModel    string  `json:"make_model" validate:"required"`
	Year         int     `json:"year"`
	Price        float64 `json:"price"`
	LocationCity string  `json:"location_city"`
	SerialNumber string  `json:"serial_number" validate:"required"`
	Description  string  `json:"description"`
}

type BikeListResponse struct {
	Bikes  []domain.Bike `json:"bikes" validate:"max=1000"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

func parsePagination(c echo.Context) (int, int, error) {
	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	limit := 1000
	offset := 0

	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l < 0 {
			return 0, 0, echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "invalid limit"})
		}
		limit = l
	}
	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil || o < 0 {
			return 0, 0, echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "invalid offset"})
		}
		offset = o
	}
	return limit, offset, nil
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
// @Failure 409 {object} map[string]string "serial number already registered"
// @Failure 500 {object} map[string]string
// @Router /bikes [post]
func (h *BikeHandler) RegisterBike(c echo.Context) error {
	userClaims, err := auth.GetClaims(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req RegisterBikeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "validation failed"})
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
		if errors.Is(err, service.ErrSerialNumberExists) {
			return c.JSON(http.StatusConflict, map[string]string{"error": "serial number already registered"})
		}
		log.Printf("RegisterBike error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return c.JSON(http.StatusCreated, bike)
}

// ListMarketplace returns only bikes with status 'for_sale' (Public)
// @Summary List public bikes for sale
// @Description Returns a list of bikes with status 'for_sale'. Sensitive fields are redacted.
// @Tags bikes
// @Produce json
// @Param limit query int false "Maximum number of bikes to return (max 1000)" default(1000)
// @Param offset query int false "Number of bikes to skip" default(0)
// @Success 200 {object} BikeListResponse
// @Failure 400 {object} map[string]string "invalid limit or offset"
// @Failure 500 {object} map[string]string
// @Router /bikes [get]
func (h *BikeHandler) ListMarketplace(c echo.Context) error {
	limit, offset, err := parsePagination(c)
	if err != nil {
		return err
	}
	bikes, total, effectiveLimit, err := h.service.ListMarketplace(c.Request().Context(), limit, offset)
	if err != nil {
		log.Printf("ListMarketplace error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
	return c.JSON(http.StatusOK, BikeListResponse{
		Bikes:  bikes,
		Total:  total,
		Limit:  effectiveLimit,
		Offset: offset,
	})
}

// ListMyBikes returns all bikes owned by the current user (Protected)
// @Summary List my bikes
// @Description Returns all bikes owned by the authenticated user with full metadata.
// @Tags bikes
// @Produce json
// @Security Bearer
// @Param limit query int false "Maximum number of bikes to return (max 1000)" default(1000)
// @Param offset query int false "Number of bikes to skip" default(0)
// @Success 200 {object} BikeListResponse
// @Failure 400 {object} map[string]string "invalid limit or offset"
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /my/bikes [get]
func (h *BikeHandler) ListMyBikes(c echo.Context) error {
	userClaims, err := auth.GetClaims(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	limit, offset, err := parsePagination(c)
	if err != nil {
		return err
	}
	bikes, total, effectiveLimit, err := h.service.ListMyBikes(c.Request().Context(), userID, limit, offset)
	if err != nil {
		log.Printf("ListMyBikes error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
	return c.JSON(http.StatusOK, BikeListResponse{
		Bikes:  bikes,
		Total:  total,
		Limit:  effectiveLimit,
		Offset: offset,
	})
}

// ListAdmin returns every bike in the system (Admin Only)
// @Summary List all bikes (Admin)
// @Description Returns every bike in the system. Strictly for users with admin role.
// @Tags admin
// @Produce json
// @Security Bearer
// @Param limit query int false "Maximum number of bikes to return (max 1000)" default(1000)
// @Param offset query int false "Number of bikes to skip" default(0)
// @Success 200 {object} BikeListResponse
// @Failure 400 {object} map[string]string "invalid limit or offset"
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/bikes [get]
func (h *BikeHandler) ListAdmin(c echo.Context) error {
	userClaims, err := auth.GetClaims(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	if userClaims.Role != "admin" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}

	limit, offset, err := parsePagination(c)
	if err != nil {
		return err
	}

	bikes, total, effectiveLimit, err := h.service.ListAdmin(c.Request().Context(), limit, offset)
	if err != nil {
		log.Printf("ListAdmin error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return c.JSON(http.StatusOK, BikeListResponse{
		Bikes:  bikes,
		Total:  total,
		Limit:  effectiveLimit,
		Offset: offset,
	})
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

	userClaims, err := auth.GetClaims(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	bike, err := h.service.GetBike(c.Request().Context(), id, userClaims.UserID, userClaims.Role)
	if err != nil {
		if errors.Is(err, service.ErrBikeNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "bike not found"})
		}
		log.Printf("GetBike error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return c.JSON(http.StatusOK, bike)
}
