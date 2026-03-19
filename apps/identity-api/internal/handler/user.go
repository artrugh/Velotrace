package handler

import (
	"context"
	"net/http"
	"os"

	"velotrace.local/auth"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/velotrace/identity-api/internal/models"
	"google.golang.org/api/idtoken"
)

type AuthGoogleRequest struct {
	Credential string `json:"credential"`
}

type AuthGoogleResponse struct {
	User  models.User `json:"user"`
	Token string      `json:"token"`
}

type UserHandler struct {
	DB *pgxpool.Pool
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
	var req AuthGoogleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.Credential == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "credential is required"})
	}

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	payload, err := idtoken.Validate(context.Background(), req.Credential, clientID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid google token", "details": err.Error()})
	}

	googleID := payload.Subject
	name, _ := payload.Claims["name"].(string)
	email, ok := payload.Claims["email"].(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "email claim missing"})
	}

	var user models.User
	err = h.DB.QueryRow(context.Background(), `
		INSERT INTO users (google_id, email, display_name, last_login, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (google_id) DO UPDATE SET
			last_login = NOW(),
			updated_at = NOW()
		RETURNING id, email, display_name, is_verified, last_login, created_at, updated_at
	`, googleID, email, name).Scan(
		&user.ID, &user.Email, &user.DisplayName, &user.IsVerified, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to process user authentication", "details": err.Error()})
	}

	token, err := auth.GenerateToken(auth.UserClaims{
		UserID: user.ID.String(),
		Email:  user.Email,
	}, os.Getenv("JWT_PRIVATE_KEY"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate session token"})
	}

	return c.JSON(http.StatusOK, AuthGoogleResponse{
		User:  user,
		Token: token,
	})
}
