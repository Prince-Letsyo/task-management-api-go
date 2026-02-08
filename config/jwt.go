package config

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type JwtSecrets struct {
	RefreshExpire time.Duration `yaml:"refresh"`
	AccessExpire  time.Duration `yaml:"access"`
	Secret        string        `yaml:"secret"`
}

type UserClaims struct {
	ID    string `json:"id"`
	First string `json:"first"`
	Last  string `json:"last"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// Config defines the config for BasicAuth middleware
type AuthConfig struct {
	// Filter defines a function to skip middleware.
	Filter func(*fiber.Ctx) bool

	// SuccessHandler defines a function which is executed for a valid token.
	SuccessHandler fiber.Handler

	// ErrorHandler defines a function which is executed for an invalid token.
	ErrorHandler fiber.ErrorHandler

	// Signing key to validate token.
	SigningKey interface{}

	// Map of signing keys to validate token with kid field usage.
	SigningKeys map[string]interface{}

	// Signing method, used to check token signing method.
	SigningMethod string

	// Context key to store user information from the token into context.
	ContextKey string

	// Claims are extendable claims data defining token content.
	Claims jwt.Claims

	// TokenLookup is a string in the form of "<source>:<name>" that is used
	// to extract token from the request.
	TokenLookup string

	// AuthScheme to be used in the Authorization header.
	AuthScheme string

	KeyFunc jwt.Keyfunc
}
