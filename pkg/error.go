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
	ErrBookNotFound        = NewAppError(fiber.StatusNotFound, "book not found")
	ErrUserNotFound        = NewAppError(fiber.StatusNotFound, "user not found")
	ErrUserSettingNotFound = NewAppError(fiber.StatusNotFound, "user setting not found")
	ErrProfileNotFound     = NewAppError(fiber.StatusNotFound, "profile not found")
	ErrInternalServer      = NewAppError(fiber.StatusInternalServerError, "internal server error")
)

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) StatusCode() int {
	return e.Code
}

func NewAppError(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

type statusCoder interface {
	StatusCode() int
}

type httpError struct {
	Statuscode int    `json:"statusCode"`
	Error      string `json:"error"`
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	// Status code defaults to 500
	code := fiber.StatusInternalServerError

	// Check for specialized error types
	var (
		fiberErr *fiber.Error
		scErr    statusCoder
	)

	if errors.As(err, &fiberErr) {
		code = fiberErr.Code
	} else if errors.As(err, &scErr) {
		code = scErr.StatusCode()
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
