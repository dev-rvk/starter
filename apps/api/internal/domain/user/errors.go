// Package user domain errors. Uses the structured domain.Error constructors
// so errors carry Kind, Entity, and Message. Validation field errors are
// handled generically by platformvalidator.ValidateAndMap — no per-field
// sentinel needed here.
package user

import "github.com/starterpack/api/internal/domain"

var (
	ErrNotFound      = domain.NotFound("user")
	ErrAlreadyExists = domain.AlreadyExists("user")
)
