// Package jwtutil provides JWT signing and verification for local auth.
// Use New() to create a JWTManager and inject it into services; no init() or
// package-level globals.
package jwtutil

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims extends the standard JWT claims with the account ID.
type Claims struct {
	jwt.RegisteredClaims
	AccountID string `json:"account_id"`
}

// JWTManager handles JWT signing and verification.
type JWTManager struct {
	secret   []byte
	duration time.Duration
}

// New creates a JWTManager with the given HMAC secret and token lifetime.
func New(secret string, duration time.Duration) *JWTManager {
	return &JWTManager{
		secret:   []byte(secret),
		duration: duration,
	}
}

// Sign creates a signed JWT for the given account ID.
func (m *JWTManager) Sign(accountID string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   accountID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.duration)),
		},
		AccountID: accountID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("jwt sign: %w", err)
	}
	return signed, nil
}

// Verify parses and validates a JWT, returning the claims if valid.
func (m *JWTManager) Verify(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(_ *jwt.Token) (any, error) {
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt verify: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("jwt verify: invalid token")
	}
	return claims, nil
}
