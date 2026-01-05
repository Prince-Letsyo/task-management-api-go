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
	ErrProfileNotFound     = errors.New("profile not found")
	ErrInternalServer      = errors.New("internal server error")
)

type httpError struct {
	Statuscode int    `json:"statusCode"`
	Error      string `json:"error"`
}

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

var (
	ErrInvalidKey      = errors.New("key is invalid")
	ErrInvalidKeyType  = errors.New("key is of invalid type")
	ErrHashUnavailable = errors.New("the requested hash function is unavailable")
)

const (
	ValidationErrorMalformed        uint32 = 1 << iota // Token is malformed
	ValidationErrorUnverifiable                        // Token could not be verified because of signing problems
	ValidationErrorSignatureInvalid                    // Signature validation failed

	ValidationErrorAudience      // AUD validation failed
	ValidationErrorExpired       // EXP validation failed
	ValidationErrorIssuedAt      // IAT validation failed
	ValidationErrorIssuer        // ISS validation failed
	ValidationErrorNotValidYet   // NBF validation failed
	ValidationErrorId            // JTI validation failed
	ValidationErrorClaimsInvalid // Generic claims validation error
)

func NewValidationError(errorText string, errorFlags uint32) *ValidationError {
	return &ValidationError{
		text:   errorText,
		Errors: errorFlags,
	}
}

type ValidationError struct {
	Inner  error  // stores the error returned by external dependencies, i.e.: KeyFunc
	Errors uint32 // bitfield.  see ValidationError... constants
	text   string // errors that do not have a valid error just have text
}

func (e ValidationError) Error() string {
	if e.Inner != nil {
		return e.Inner.Error()
	} else if e.text != "" {
		return e.text
	} else {
		return "token is invalid"
	}
}

func (e *ValidationError) Valid() bool {
	return e.Errors == 0
}
