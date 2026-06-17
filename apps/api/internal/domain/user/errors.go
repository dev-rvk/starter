package user

import (
	"fmt"

	"github.com/starterpack/api/internal/domain"
)

// Domain-level errors. These wrap the shared sentinels from the domain package
// so the HTTP adapter can map them by category (not-found, conflict, validation)
// without importing every individual domain.
var (
	ErrInvalidUsername = fmt.Errorf("username must be 2-6 characters, letters/digits/underscores: %w", domain.ErrValidation)
	ErrInvalidEmail    = fmt.Errorf("invalid email address: %w", domain.ErrValidation)
	ErrNotFound        = fmt.Errorf("user %w", domain.ErrNotFound)
	ErrAlreadyExists   = fmt.Errorf("user %w", domain.ErrAlreadyExists)
)
