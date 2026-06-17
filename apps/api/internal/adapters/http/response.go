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

// mapDomainError translates domain errors into HTTP statuses. Because every
// domain wraps the shared sentinels (domain.ErrNotFound, domain.ErrValidation,
// etc.), this function works for any domain without importing it.
func mapDomainError(err error) (int, string) {
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

