package validator

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func Struct(s interface{}) error {
	return validate.Struct(s)
}

func Field(field interface{}, tag string) error {
	return validate.Var(field, tag)
}

func Errors(err error) []string {
	if err == nil {
		return nil
	}

	var errors []string
	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, formatError(err))
	}
	return errors
}

func formatError(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return err.Field() + " is required"
	case "email":
		return err.Field() + " must be a valid email address"
	case "min":
		return err.Field() + " must be at least " + err.Param() + " characters"
	case "max":
		return err.Field() + " must be at most " + err.Param() + " characters"
	case "oneof":
		return err.Field() + " must be one of: " + err.Param()
	default:
		return err.Field() + " is invalid"
	}
}
