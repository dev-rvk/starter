// Package domain defines shared domain-level error types. Individual domains
// (user, todo, …) use the convenience constructors so the HTTP adapter can map
// any domain error to the correct status code via errors.As() without importing
// every domain package.
package domain

import "errors"

// ErrorKind categorizes domain errors for the HTTP layer.
type ErrorKind int

const (
	KindUnknown       ErrorKind = iota // 0 — zero value; never set intentionally
	KindNotFound                       // 1 — 404
	KindAlreadyExists                  // 2 — 409
	KindValidation                     // 3 — 422
)

// Sentinel errors for backward-compatible errors.Is() checks. The structured
// Error type wraps these so both errors.Is() and errors.As() work.
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrValidation    = errors.New("validation failed")
)

// Error is the structured domain error. It carries a Kind for HTTP mapping,
// an optional Field for validation errors, and a human-readable Message.
// It wraps the corresponding sentinel so errors.Is() still works.
type Error struct {
	Kind    ErrorKind
	Entity  string // "user", "todo", etc.
	Field   string // non-empty for validation errors
	Message string
	err     error // wrapped sentinel — access via Unwrap()
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.err }

// NotFound creates a structured not-found error for the given entity.
func NotFound(entity string) *Error {
	return &Error{
		Kind:    KindNotFound,
		Entity:  entity,
		Message: entity + " not found",
		err:     ErrNotFound,
	}
}

// AlreadyExists creates a structured conflict error for the given entity.
func AlreadyExists(entity string) *Error {
	return &Error{
		Kind:    KindAlreadyExists,
		Entity:  entity,
		Message: entity + " already exists",
		err:     ErrAlreadyExists,
	}
}

// ValidationError creates a structured validation error for a specific field.
func ValidationError(entity, field, reason string) *Error {
	return &Error{
		Kind:    KindValidation,
		Entity:  entity,
		Field:   field,
		Message: reason,
		err:     ErrValidation,
	}
}
