package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// phoneRegex allows digits, +, -, spaces, parentheses
var phoneRegex = regexp.MustCompile(`^[0-9+\-\s()]+$`)

func init() {
	err := validate.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		if val == "" {
			return true
		}
		return phoneRegex.MatchString(val)
	})
	if err != nil {
		panic("failed to register phone validator: " + err.Error())
	}
}

// Struct validates a struct using go-playground/validator tags.
// Returns a user-friendly error message or nil if valid.
func Struct(s interface{}) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return fmt.Errorf("validation error")
	}

	messages := make([]string, 0, len(validationErrors))
	for _, e := range validationErrors {
		messages = append(messages, formatFieldError(e))
	}

	return fmt.Errorf("%s", strings.Join(messages, "; "))
}

func formatFieldError(e validator.FieldError) string {
	field := toSnakeCase(e.Field())

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
	case "uuid4", "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "phone":
		return fmt.Sprintf("%s must contain only digits, +, -, spaces, or parentheses", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, e.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, e.Param())
	case "gte":
		return fmt.Sprintf("%s must be at least %s", field, e.Param())
	case "lte":
		return fmt.Sprintf("%s must be at most %s", field, e.Param())
	default:
		return fmt.Sprintf("%s is invalid (%s)", field, e.Tag())
	}
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
