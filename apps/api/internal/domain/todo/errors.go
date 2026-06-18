package todo

import (
	"fmt"

	"github.com/starterpack/api/internal/domain"
)

var (
	ErrNotFound     = fmt.Errorf("todo %w", domain.ErrNotFound)
    ErrInvalidTitle = fmt.Errorf("todo title must be between 1 and 50 characters: %w", domain.ErrValidation)
)