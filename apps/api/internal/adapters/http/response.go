package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/starterpack/api/internal/domain"
)

// ErrorResponse is the standard error envelope returned by the API.
type ErrorResponse struct {
	Error string `json:"error"`
}

func respondError(c *gin.Context, status int, msg string) {
	c.JSON(status, ErrorResponse{Error: msg})
}

// mapDomainError translates domain errors into HTTP statuses. It prefers the
// structured domain.Error type (via errors.As) and falls back to sentinel
// errors.Is checks for backward compatibility.
func mapDomainError(err error) (int, string) {
	// Prefer the structured error type.
	var domErr *domain.Error
	if errors.As(err, &domErr) {
		switch domErr.Kind {
		case domain.KindNotFound:
			return http.StatusNotFound, domErr.Message
		case domain.KindAlreadyExists:
			return http.StatusConflict, domErr.Message
		case domain.KindValidation:
			return http.StatusUnprocessableEntity, domErr.Message
		default:
			return http.StatusInternalServerError, "internal server error"
		}
	}

	// Fallback: sentinel errors (backward compat).
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, domain.ErrAlreadyExists):
		return http.StatusConflict, err.Error()
	case errors.Is(err, domain.ErrValidation):
		return http.StatusUnprocessableEntity, err.Error()
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

