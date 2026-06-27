// Package user is the domain core for users.
package user

import (
	"time"
)

// User is the domain entity.
type User struct {
	ID        string    `validate:"required"`
	Username  string    `validate:"required,username"`
	Email     string    `validate:"required,email"`
	CreatedAt time.Time `validate:"required"`
	UpdatedAt time.Time `validate:"required"`
}

// New builds a User.
func New(id, username, email string, now time.Time) *User {
	return &User{
		ID:        id,
		Username:  username,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
