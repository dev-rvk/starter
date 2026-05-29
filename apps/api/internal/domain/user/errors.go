package user

import "errors"

// Domain-level errors. Adapters translate these into transport-specific
// representations (e.g. HTTP status codes) — the domain stays transport-agnostic.
var (
	ErrInvalidUsername = errors.New("username must be 2-6 characters and contain only letters, numbers or underscores")
	ErrInvalidEmail    = errors.New("invalid email address")
	ErrNotFound        = errors.New("user not found")
	ErrAlreadyExists   = errors.New("user already exists")
)
