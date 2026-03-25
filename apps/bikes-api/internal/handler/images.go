package handler

import (
	"context"
	"fmt"
	"net/http"

	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/velotrace/bikes-api/internal/models"
	"github.com/velotrace/bikes-api/internal/platform"
	"velotrace.local/auth"
)

type ImageHandler struct {
	DB      *pgxpool.Pool
	Storage *platform.Storage
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
	bikeID := c.Param("id")
	if _, err := uuid.Parse(bikeID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid bike id"})
	}

	var req UploadURLRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	timestamp := time.Now().Unix()
	objectKey := fmt.Sprintf("bikes/%s/%d_%s", bikeID, timestamp, req.Filename)

	uploadURL, err := h.Storage.GetPresignedPutURL(c.Request().Context(), objectKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate upload URL"})
	}

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
// @Failure 500 {object} map[string]string
// @Router /bikes/{id}/images/confirm [post]
func (h *ImageHandler) ConfirmUpload(c echo.Context) error {
	bikeIDStr := c.Param("id")
	bikeID, err := uuid.Parse(bikeIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid bike id"})
	}

	var req ConfirmUploadRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	// Verify the user is the owner of the bike
	userClaims := c.Get("user").(*auth.UserClaims)
	userID, _ := uuid.Parse(userClaims.UserID)

	var ownerID uuid.UUID
	err = h.DB.QueryRow(context.Background(), "SELECT current_owner_id FROM bikes WHERE id = $1", bikeID).Scan(&ownerID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "bike not found"})
	}
	if ownerID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "not the owner of this bike"})
	}

	objectKey := req.ObjectKey

	// Check if first image
	var count int
	err = h.DB.QueryRow(context.Background(), "SELECT COUNT(*) FROM bike_images WHERE bike_id = $1", bikeID).Scan(&count)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to check image count"})
	}

	isPrimary := count == 0
	_, err = h.DB.Exec(context.Background(), `
		INSERT INTO bike_images (bike_id, object_key, is_primary)
		VALUES ($1, $2, $3)
	`, bikeID, objectKey, isPrimary)
	if err != nil {
		fmt.Printf("DATABASE ERROR: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save bike image", "details": err.Error()})
	}
	img := models.BikeImage{ObjectKey: objectKey}
	img.PopulateURL()

	return c.JSON(http.StatusCreated, map[string]string{
		"status": "success",
		"url":    img.URL,
	})
}
