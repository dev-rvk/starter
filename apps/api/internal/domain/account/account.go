// Package account is the domain core for authentication credentials.
// It stores email + password hash, separate from the user domain (app profiles).
package account

import (
	"time"
)

// Account is the domain entity for authentication credentials.
type Account struct {
	ID           string    `validate:"required"`
	Email        string    `validate:"required,email"`
	PasswordHash string    `validate:"required"`
	CreatedAt    time.Time `validate:"required"`
	UpdatedAt    time.Time `validate:"required"`
}

// New builds an Account. The caller must hash the password before calling this.
func New(id string, email string, passwordHash string, now time.Time) *Account {
	return &Account{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
