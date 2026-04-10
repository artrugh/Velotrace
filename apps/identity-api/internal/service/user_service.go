package service

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/velotrace/identity-api/internal/domain"
	"google.golang.org/api/idtoken"
	"velotrace.local/auth"
)

var (
	ErrInvalidGoogleToken    = errors.New("invalid google token")
	ErrEmailClaimMissing     = errors.New("email claim missing")
	ErrMissingClientID       = errors.New("GOOGLE_CLIENT_ID is not configured")
	ErrFailedToGenerateToken = errors.New("failed to generate session token")
	ErrMissingPrivateKey     = errors.New("JWT_PRIVATE_KEY is not configured")
)

type UserRepository interface {
	UpsertByGoogleID(ctx context.Context, googleID, email, displayName string) (*domain.User, error)
}

type TokenValidator interface {
	Validate(ctx context.Context, idToken string, audience string) (*idtoken.Payload, error)
}

type TokenGenerator interface {
	GenerateToken(claims auth.UserClaims) (string, error)
}

type googleTokenValidator struct{}

func (v *googleTokenValidator) Validate(ctx context.Context, idToken string, audience string) (*idtoken.Payload, error) {
	return idtoken.Validate(ctx, idToken, audience)
}

type UserService interface {
	AuthGoogle(ctx context.Context, credential string) (*domain.User, string, error)
}

type userService struct {
	repo        UserRepository
	authManager TokenGenerator
	validator   TokenValidator
}

func NewUserService(repo UserRepository, authManager TokenGenerator) UserService {
	return &userService{
		repo:        repo,
		authManager: authManager,
		validator:   &googleTokenValidator{},
	}
}

func (s *userService) AuthGoogle(ctx context.Context, credential string) (*domain.User, string, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		return nil, "", ErrMissingClientID
	}

	payload, err := s.validator.Validate(ctx, credential, clientID)
	if err != nil {
		return nil, "", ErrInvalidGoogleToken
	}

	googleID := payload.Subject
	name, _ := payload.Claims["name"].(string)
	email, ok := payload.Claims["email"].(string)
	if !ok {
		return nil, "", ErrEmailClaimMissing
	}

	user, err := s.repo.UpsertByGoogleID(ctx, googleID, email, name)
	if err != nil {
		return nil, "", err
	}

	token, err := s.authManager.GenerateToken(auth.UserClaims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})
	if err != nil {
		return nil, "", ErrFailedToGenerateToken
	}

	return user, token, nil
}
