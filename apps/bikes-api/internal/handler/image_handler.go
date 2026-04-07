package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/velotrace/bikes-api/internal/service"
	"velotrace.local/auth"
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
	bikeIDStr := c.Param("id")
	bikeID, err := uuid.Parse(bikeIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid bike id"})
	}

	var req UploadURLRequest
	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	err = c.Validate(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing required fields", "details": err.Error()})
	}

	uploadURL, objectKey, err := h.service.GetUploadURL(c.Request().Context(), bikeID, req.Filename)
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
	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	err = c.Validate(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing required fields", "details": err.Error()})
	}

	userClaims := c.Get("user").(*auth.UserClaims)
	userID, err := uuid.Parse(userClaims.UserID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user token"})
	}

	url, err := h.service.ConfirmUpload(c.Request().Context(), bikeID, userID, req.ObjectKey)
	if err != nil {
		if err.Error() == "bike not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		if err.Error() == "not the owner of this bike" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save bike image", "details": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"status": "success",
		"url":    url,
	})
}
