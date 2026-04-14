package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"velotrace.local/logger"
)

type UserClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

var (
	ErrParseRSAPrivateKey      = errors.New("failed to parse RSA private key")
	ErrParseRSAPublicKey       = errors.New("failed to parse RSA public key")
	ErrPrivateKeyNotConfigured = errors.New("private key not configured")
	ErrPublicKeyNotConfigured  = errors.New("public key not configured")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrInvalidToken            = errors.New("invalid token")
	ErrMissingAuthClaims       = errors.New("missing authentication claims")
	ErrInvalidAuthClaimsFormat = errors.New("invalid authentication claims format")
)

func NewTokenManager(privateKeyPEM, publicKeyPEM string) (*TokenManager, error) {
	var privKey *rsa.PrivateKey
	var pubKey *rsa.PublicKey
	var err error

	if privateKeyPEM != "" {
		privKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrParseRSAPrivateKey, err)
		}
	}

	if publicKeyPEM != "" {
		pubKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyPEM))
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrParseRSAPublicKey, err)
		}
	}

	return &TokenManager{
		privateKey: privKey,
		publicKey:  pubKey,
	}, nil
}

func (m *TokenManager) GenerateToken(claims UserClaims) (string, error) {
	if m.privateKey == nil {
		return "", ErrPrivateKeyNotConfigured
	}

	now := time.Now()
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(24 * time.Hour))

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(m.privateKey)
}

func (m *TokenManager) ValidateToken(tokenStr string) (*UserClaims, error) {
	if m.publicKey == nil {
		return nil, ErrPublicKeyNotConfigured
	}

	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedSigningMethod, token.Header["alg"])
		}
		return m.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrInvalidToken
}

func (m *TokenManager) JWTGuard() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			l := logger.FromContext(c).With("component", "auth")

			p := c.Request().URL.Path
			if p == "/favicon.ico" || p == "/" {
				return next(c)
			}

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				l.Warn("missing authorization header")
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				l.Warn("auth failed: invalid authorization header format")
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			claims, err := m.ValidateToken(parts[1])
			if err != nil {
				l.Error("auth failed: token validation", "err", err)
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			l.Debug("validating token")
			c.Set("user", claims)
			return next(c)
		}
	}
}

func (m *TokenManager) OptionalJWTGuard() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return next(c)
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			claims, err := m.ValidateToken(parts[1])
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			c.Set("user", claims)
			return next(c)
		}
	}
}

func GetClaims(c echo.Context) (*UserClaims, error) {
	raw := c.Get("user")
	if raw == nil {
		return nil, ErrMissingAuthClaims
	}
	claims, ok := raw.(*UserClaims)
	if !ok {
		return nil, ErrInvalidAuthClaimsFormat
	}
	return claims, nil
}
