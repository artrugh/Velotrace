package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
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

func NewTokenManager(privateKeyPEM, publicKeyPEM string) (*TokenManager, error) {
	var privKey *rsa.PrivateKey
	var pubKey *rsa.PublicKey
	var err error

	if privateKeyPEM != "" {
		privKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
		if err != nil {
			return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
		}
	}

	if publicKeyPEM != "" {
		pubKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyPEM))
		if err != nil {
			return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
		}
	}

	return &TokenManager{
		privateKey: privKey,
		publicKey:  pubKey,
	}, nil
}

func (m *TokenManager) GenerateToken(claims UserClaims) (string, error) {
	if m.privateKey == nil {
		return "", errors.New("private key not configured")
	}

	now := time.Now()
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(24 * time.Hour))

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(m.privateKey)
}

func (m *TokenManager) ValidateToken(tokenStr string) (*UserClaims, error) {
	if m.publicKey == nil {
		return nil, errors.New("public key not configured")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func (m *TokenManager) JWTGuard() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				log.Printf("[Internal Auth Error]: %v\n", "missing authorization header")
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Printf("[Internal Auth Error]: %v\n", "invalid authorization format")
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			claims, err := m.ValidateToken(parts[1])
			if err != nil {
				log.Printf("[Validation Error]: %v\n", err)
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

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
		return nil, errors.New("missing authentication claims")
	}
	claims, ok := raw.(*UserClaims)
	if !ok {
		return nil, errors.New("invalid authentication claims format")
	}
	return claims, nil
}
