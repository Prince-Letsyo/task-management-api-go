package pkg

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Validateinstance struct {
	Vi *validator.Validate
}

var VI Validateinstance = Validateinstance{
	Vi: validator.New(validator.WithRequiredStructEnabled()),
}

type validationError struct {
	Namespace       string `json:"namespace"` // can differ when a custom TagNameFunc is registered or
	Field           string `json:"field"`     // by passing alt name to ReportError like below
	StructNamespace string `json:"structNamespace"`
	StructField     string `json:"structField"`
	Tag             string `json:"tag"`
	ActualTag       string `json:"actualTag"`
	Kind            string `json:"kind"`
	Type            string `json:"type"`
	Value           string `json:"value"`
	Param           string `json:"param"`
	Message         string `json:"message"`
}

func validateStruct(value interface{}) error {
	if validErr := VI.Vi.Struct(value); validErr != nil {
		for _, err := range validErr.(validator.ValidationErrors) {
			e := validationError{
				Namespace:       err.Namespace(),
				Field:           err.Field(),
				StructNamespace: err.StructNamespace(),
				StructField:     err.StructField(),
				Tag:             err.Tag(),
				ActualTag:       err.ActualTag(),
				Kind:            fmt.Sprintf("%v", err.Kind()),
				Type:            fmt.Sprintf("%v", err.Type()),
				Value:           fmt.Sprintf("%v", err.Value()),
				Param:           err.Param(),
				Message:         err.Error(),
			}

			indent, err := json.MarshalIndent(e, "", "  ")
			if err != nil {
				fmt.Println(err)
				panic(err)
			}

			fmt.Println(string(indent))

		}
		return validErr
	}
	return nil
}

func CustomQueryParser(c *fiber.Ctx, value interface{}) error {
	if err := c.QueryParser(value); err != nil {
		return err
	}
	return validateStruct(value)
}

func CustomBodyParser(c *fiber.Ctx, value any) error {
	if err := c.BodyParser(value); err != nil {
		return err
	}
	return validateStruct(value)
}
