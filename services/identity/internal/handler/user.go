package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type SignUpRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

func SignUp(c echo.Context) error {
	var req SignUpRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "SignUp placeholder - user received",
		"data":    req,
	})
}
