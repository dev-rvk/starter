// Package user is the domain core for users. It has no dependencies on
// frameworks, transport or persistence — only the standard library. Validation
// lives in the value-object constructors, so an invalid User cannot exist.
package user

import (
	"strings"
	"time"
)

// Username is a validated value object: 2-6 characters, letters/digits/underscore.
type Username struct{ value string }

// NewUsername validates and constructs a Username. Mirror these rules in the
// frontend zod schema (packages/design-system auth forms).
func NewUsername(raw string) (Username, error) {
	v := strings.TrimSpace(raw)
	if len(v) < 2 || len(v) > 6 {
		return Username{}, ErrInvalidUsername
	}
	for _, r := range v {
		isLower := r >= 'a' && r <= 'z'
		isUpper := r >= 'A' && r <= 'Z'
		isDigit := r >= '0' && r <= '9'
		if !(isLower || isUpper || isDigit || r == '_') {
			return Username{}, ErrInvalidUsername
		}
	}
	return Username{value: v}, nil
}

func (u Username) String() string { return u.value }

// Email is a validated value object holding a normalized (lower-cased) address.
type Email struct{ value string }

// NewEmail performs a deliberately simple structural validation.
func NewEmail(raw string) (Email, error) {
	v := strings.ToLower(strings.TrimSpace(raw))
	at := strings.IndexByte(v, '@')
	if at <= 0 || at >= len(v)-1 {
		return Email{}, ErrInvalidEmail
	}
	if !strings.Contains(v[at+1:], ".") || strings.Contains(v, " ") {
		return Email{}, ErrInvalidEmail
	}
	return Email{value: v}, nil
}

func (e Email) String() string { return e.value }

// User is the domain entity.
type User struct {
	ID        string
	Username  Username
	Email     Email
	CreatedAt time.Time
	UpdatedAt time.Time
}

// New builds a User from already-validated value objects.
func New(id string, username Username, email Email, now time.Time) *User {
	return &User{
		ID:        id,
		Username:  username,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
