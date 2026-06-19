// Package validator provides a shared go-playground/validator instance with
// project-specific custom rules. Use New() to create a Validator and inject
// it into services; do not call package-level functions or rely on init().
package validator

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/starterpack/api/internal/domain"
)

// usernameRegex matches 2–6 alphanumeric characters or underscores.
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{2,6}$`)

// Validator wraps go-playground/validator with domain-aware error mapping.
type Validator struct {
	v *validator.Validate
}

// New creates a Validator with all custom validation rules registered.
func New() *Validator {
	v := validator.New()
	_ = v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		return usernameRegex.MatchString(fl.Field().String())
	})
	return &Validator{v: v}
}

// ValidateAndMap validates a struct and maps the first field failure to a
// structured domain.Error, eliminating per-field sentinel boilerplate in
// every service method.
func (vl *Validator) ValidateAndMap(entity string, s any) error {
	err := vl.v.Struct(s)
	if err == nil {
		return nil
	}
	var valErrs validator.ValidationErrors
	if errors.As(err, &valErrs) {
		fe := valErrs[0]
		return domain.ValidationError(
			entity,
			fe.Field(),
			fmt.Sprintf("%s failed on '%s' rule", fe.Field(), fe.Tag()),
		)
	}
	return err
}
