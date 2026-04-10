package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateKeyPair(t *testing.T) (string, string) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(privateKeyPEM), string(publicKeyPEM)
}

func TestTokenManager_Lifecycle(t *testing.T) {
	priv, pub := generateKeyPair(t)
	manager, err := NewTokenManager(priv, pub)
	require.NoError(t, err)

	claims := UserClaims{
		UserID: "user-123",
		Email:  "test@example.com",
		Role:   "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	// 1. Generate
	token, err := manager.GenerateToken(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 2. Validate
	parsedClaims, err := manager.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, claims.UserID, parsedClaims.UserID)
	assert.Equal(t, claims.Email, parsedClaims.Email)
	assert.Equal(t, claims.Role, parsedClaims.Role)
}

func TestTokenManager_InvalidKey(t *testing.T) {
	_, err := NewTokenManager("invalid-key", "")
	assert.Error(t, err)
}

func TestTokenManager_MissingPrivateKey(t *testing.T) {
	_, pub := generateKeyPair(t)
	manager, err := NewTokenManager("", pub)
	require.NoError(t, err)

	_, err = manager.GenerateToken(UserClaims{})
	assert.EqualError(t, err, "private key not configured")
}
