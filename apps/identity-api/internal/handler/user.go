package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/velotrace/identity-api/internal/domain"
	"github.com/velotrace/identity-api/internal/service"
	"velotrace.local/logger"
)

type AuthGoogleRequest struct {
	Credential string `json:"credential" validate:"required"`
}

type AuthGoogleResponse struct {
	User  domain.User `json:"user"`
	Token string      `json:"token"`
}

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// AuthGoogle handles Google OAuth login
// @Summary Google Login
// @Description Authenticate a user using a Google ID token from GSI
// @Tags auth
// @Accept json
// @Produce json
// @Param request body AuthGoogleRequest true "Google Credential"
// @Success 200 {object} AuthGoogleResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/google [post]
func (h *UserHandler) AuthGoogle(c echo.Context) error {
	l := logger.FromContext(c)

	var req AuthGoogleRequest
	if err := c.Bind(&req); err != nil {
		l.Error("json bind failure", "err", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.Credential == "" {
		l.Warn("missing google credential in request")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "credential is required"})
	}

	user, token, err := h.userService.AuthGoogle(c.Request().Context(), req.Credential)
	if err != nil {
		if errors.Is(err, service.ErrInvalidGoogleToken) || errors.Is(err, service.ErrEmailClaimMissing) {
			l.Warn("google auth rejected", "err", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		l.Error("AuthGoogle service failure", "err", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	l.Info("user authenticated via google", "user_id", user.ID)
	return c.JSON(http.StatusOK, AuthGoogleResponse{
		User:  *user,
		Token: token,
	})
}
