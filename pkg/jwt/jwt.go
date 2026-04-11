package jwt

import (
	"errors"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

// Claims holds the JWT payload.
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	gojwt.RegisteredClaims
}

// Manager handles JWT signing and verification.
type Manager struct {
	secret      []byte
	expiryHours int
}

// NewManager creates a JWT manager.
func NewManager(secret string, expiryHours int) *Manager {
	return &Manager{
		secret:      []byte(secret),
		expiryHours: expiryHours,
	}
}

// Generate creates a signed JWT token for the given user.
func (m *Manager) Generate(userID, username, role string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: gojwt.RegisteredClaims{
			IssuedAt:  gojwt.NewNumericDate(time.Now()),
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(time.Duration(m.expiryHours) * time.Hour)),
			Issuer:    "salary-advance-service",
		},
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Validate parses and validates a token string, returning the claims on success.
func (m *Manager) Validate(tokenStr string) (*Claims, error) {
	token, err := gojwt.ParseWithClaims(tokenStr, &Claims{}, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}
