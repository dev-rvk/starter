// Package domain defines shared domain-level error sentinels. Individual domains
// (user, todo, …) wrap these with context so the HTTP adapter can map any domain
// error to the correct status code without importing every domain package.
package domain

import "errors"

var (
	// ErrNotFound indicates the requested entity does not exist.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists indicates a uniqueness constraint violation.
	ErrAlreadyExists = errors.New("already exists")

	// ErrValidation indicates input failed business-rule validation.
	ErrValidation = errors.New("validation failed")
)
