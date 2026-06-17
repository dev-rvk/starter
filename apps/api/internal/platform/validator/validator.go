package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
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
