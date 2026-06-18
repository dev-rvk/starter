package validator

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/starterpack/api/internal/domain"
)

var (
	// usernameRegex matches 2-6 alphanumeric characters or underscores
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{2,6}$`)
	validate      *validator.Validate
)

func init() {
	validate = validator.New()

	// Register custom validator tag "username"
	_ = validate.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		return usernameRegex.MatchString(fl.Field().String())
	})
}

// ValidateStruct runs struct tag validation using the initialized validator.
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// Instance returns the raw validator instance for extension.
func Instance() *validator.Validate {
	return validate
}

// ValidateAndMap validates a struct and maps the first field failure to a
// structured domain.Error. This eliminates the per-field sentinel boilerplate
// that would otherwise be needed in every service method.
func ValidateAndMap(entity string, s interface{}) error {
	err := ValidateStruct(s)
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
