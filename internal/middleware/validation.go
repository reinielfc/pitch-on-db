package middleware

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func formatValidationErrors(errs validator.ValidationErrors) []map[string]string {
	out := make([]map[string]string, len(errs))
	for i, e := range errs {
		out[i] = map[string]string{
			"field":   e.Field(),
			"rule":    e.Tag(),
			"message": humanMessage(e),
		}
	}
	return out
}

func humanMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", e.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", e.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", e.Field(), e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", e.Field(), e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", e.Field(), e.Param())
	default:
		return fmt.Sprintf("%s failed %s validation", e.Field(), e.Tag())
	}
}
