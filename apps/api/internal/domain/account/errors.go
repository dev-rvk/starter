// Package account domain errors. Uses the structured domain.Error constructors
// so errors carry Kind, Entity, and Message.
package account

import "github.com/starterpack/api/internal/domain"

var (
	ErrNotFound      = domain.NotFound("account")
	ErrAlreadyExists = domain.AlreadyExists("account")
)
