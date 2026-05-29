package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	userdomain "github.com/starterpack/api/internal/domain/user"
)

// ErrorResponse is the standard error envelope returned by the API.
type ErrorResponse struct {
	Error string `json:"error"`
}

func respondError(c *gin.Context, status int, msg string) {
	c.JSON(status, ErrorResponse{Error: msg})
}

// mapDomainError translates transport-agnostic domain errors into HTTP statuses.
func mapDomainError(err error) (int, string) {
	switch {
	case errors.Is(err, userdomain.ErrNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, userdomain.ErrAlreadyExists):
		return http.StatusConflict, err.Error()
	case errors.Is(err, userdomain.ErrInvalidUsername), errors.Is(err, userdomain.ErrInvalidEmail):
		return http.StatusUnprocessableEntity, err.Error()
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
