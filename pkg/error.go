package pkg

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

var (
	ErrValidation = errors.New("validation error")

	ErrPageLimit = fmt.Errorf("invalid page limit [max=%v]", MaxPageLimit)

	ErrBookName            = errors.New("invalid name")
	ErrBookNotFound        = errors.New("book not found")
	ErrUserNotFound        = errors.New("user not found")
	ErrUserSettingNotFound = errors.New("user setting not found")
)

type httpError struct {
	Statuscode int    `json:"statusCode"`
	Error      string `json:"error"`
}

// ErrorHandler is used to catch error thrown inside the routes by ctx.Next(err)
func ErrorHandler(c *fiber.Ctx, err error) error {
	// Statuscode defaults to 500
	code := fiber.StatusInternalServerError

	// Check if it's an fiber.Error type
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(&httpError{
		Statuscode: code,
		Error:      err.Error(),
	})
}
