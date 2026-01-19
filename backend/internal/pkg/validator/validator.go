package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	v := validator.New()

	// Use JSON tag names in error messages
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{validate: v}
}

func (v *Validator) Validate(i interface{}) map[string]string {
	err := v.validate.Struct(i)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)
	for _, err := range err.(validator.ValidationErrors) {
		field := err.Field()
		switch err.Tag() {
		case "required":
			errors[field] = field + " is required"
		case "email":
			errors[field] = field + " must be a valid email address"
		case "min":
			errors[field] = field + " must be at least " + err.Param() + " characters"
		case "max":
			errors[field] = field + " must be at most " + err.Param() + " characters"
		case "alphanum":
			errors[field] = field + " must contain only alphanumeric characters"
		case "url":
			errors[field] = field + " must be a valid URL"
		case "uuid":
			errors[field] = field + " must be a valid UUID"
		case "oneof":
			errors[field] = field + " must be one of: " + err.Param()
		case "gt":
			errors[field] = field + " must be greater than " + err.Param()
		case "gte":
			errors[field] = field + " must be greater than or equal to " + err.Param()
		case "gtfield":
			errors[field] = field + " must be greater than " + err.Param()
		case "gtefield":
			errors[field] = field + " must be greater than or equal to " + err.Param()
		case "numeric":
			errors[field] = field + " must be a valid number"
		default:
			errors[field] = field + " is invalid"
		}
	}

	return errors
}

// ValidateVar validates a single variable
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}
