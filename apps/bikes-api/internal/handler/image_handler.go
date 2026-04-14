package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/velotrace/bikes-api/internal/domain"
	"github.com/velotrace/bikes-api/internal/service"
	"velotrace.local/auth"
	"velotrace.local/logger"
)

type ImageHandler struct {
	service *service.ImageService
}

func NewImageHandler(service *service.ImageService) *ImageHandler {
	return &ImageHandler{service: service}
}

type UploadURLRequest struct {
	Filename string `json:"filename" validate:"required"`
}

type UploadURLResponse struct {
	UploadURL string `json:"upload_url"`
	ObjectKey string `json:"object_key"`
}

type ConfirmUploadRequest struct {
	ObjectKey string `json:"object_key" validate:"required"`
}

// GetUploadURL generates a presigned PUT URL for image upload
// @Summary Get presigned upload URL
// @Description Generates a unique object key and a presigned URL valid for 15 minutes
// @Tags images
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Bike ID"
// @Param request body UploadURLRequest true "Upload request"
// @Success 200 {object} UploadURLResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /bikes/{id}/upload-url [post]
func (h *ImageHandler) GetUploadURL(c echo.Context) error {
	l := logger.FromContext(c)

	bikeIDStr := c.Param("id")
	bikeID, err := uuid.Parse(bikeIDStr)
	if err != nil {
		l.Warn("invalid bike id param", "input", bikeIDStr)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid bike id"})
	}

	var req UploadURLRequest
	if err := c.Bind(&req); err != nil {
		l.Warn("json bind failure", "err", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		l.Warn("validation failure", "err", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "validation failed"})
	}

	uploadURL, objectKey, err := h.service.GetUploadURL(c.Request().Context(), bikeID, req.Filename)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidFilename) {
			l.Info("invalid filename rejected", "filename", req.Filename)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid filename"})
		}
		l.Error("GetUploadURL service failure", "err", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	l.Info("presigned upload url generated", "bike_id", bikeID, "object_key", objectKey)
	return c.JSON(http.StatusOK, UploadURLResponse{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
	})
}

// ConfirmUpload creates a database record for the uploaded image
// @Summary Confirm image upload
// @Description Creates a bike_image record. Sets as primary if it's the first image.
// @Tags images
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Bike ID"
// @Param request body ConfirmUploadRequest true "Confirm upload request"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string "forbidden"
// @Failure 404 {object} map[string]string "bike not found"
// @Failure 500 {object} map[string]string
// @Router /bikes/{id}/images/confirm [post]
func (h *ImageHandler) ConfirmUpload(c echo.Context) error {
	l := logger.FromContext(c)

	bikeIDStr := c.Param("id")
	bikeID, err := uuid.Parse(bikeIDStr)
	if err != nil {
		l.Warn("invalid bike id param", "input", bikeIDStr)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid bike id"})
	}

	var req ConfirmUploadRequest
	if err := c.Bind(&req); err != nil {
		l.Warn("json bind failure", "err", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		l.Warn("validation failure", "err", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "validation failed"})
	}

	userClaims, err := auth.GetClaims(c)
	if err != nil {
		l.Error("auth claims missing", "err", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		l.Error("failed to parse userID from claims", "err", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	url, err := h.service.ConfirmUpload(c.Request().Context(), bikeID, userID, req.ObjectKey)
	if err != nil {
		if errors.Is(err, domain.ErrBikeNotFound) {
			l.Info("bike not found during image confirm", "bike_id", bikeID)
			return c.JSON(http.StatusNotFound, map[string]string{"error": "bike not found"})
		}
		if errors.Is(err, domain.ErrNotOwner) {
			l.Warn("forbidden: user is not owner", "user_id", userID, "bike_id", bikeID)
			return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
		}
		l.Error("ConfirmUpload service failure", "err", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	l.Info("image upload confirmed", "bike_id", bikeID, "url", url)
	return c.JSON(http.StatusCreated, map[string]string{
		"status": "success",
		"url":    url,
	})
}
