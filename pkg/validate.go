// Package pkg dsf
package pkg

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// 1. Define the structured error format
type ErrorResponse struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

// 2. Initialize Validator with JSON tag support
type Validateinstance struct {
	Vi *validator.Validate
}

var VI = Validateinstance{
	Vi: func() *validator.Validate {
		v := validator.New(validator.WithRequiredStructEnabled())

		// Returns the name from the json:"" tag instead of the Struct field name
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		return v
	}(),
}

func ValidateStruct(value interface{}) []ErrorResponse {
	var errors []ErrorResponse
	err := VI.Vi.Struct(value)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var msg string

			// Custom message mapping based on the tag
			switch err.Tag() {
			case "required":
				msg = fmt.Sprintf("%s is required", err.Field())
			case "email":
				msg = fmt.Sprintf("%s must be a valid email address", err.Field())
			case "min":
				msg = fmt.Sprintf("%s must be at least %s characters", err.Field(), err.Param())
			case "unique":
				msg = fmt.Sprintf("%s is already in use", err.Field())
			default:
				msg = fmt.Sprintf("field %s failed validation on %s", err.Field(), err.Tag())
			}

			errors = append(errors, ErrorResponse{
				Field:   err.Field(),
				Tag:     err.Tag(),
				Message: msg,
			})
		}
	}
	return errors
}
