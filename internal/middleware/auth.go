// Package middleware implements authentication and authorization middleware for HTTP handlers.
package middleware

import (
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/gofiber/fiber/v2"
)

type (
	IAuthenticateMiddleWare interface {
		Authenticate() fiber.Handler
		AuthenticateAdmin(*fiber.Ctx) error
	}

	IAuthenticate interface {
		Auth() fiber.Handler
		AuthAdmin(*fiber.Ctx) error
	}
)

type AuthAuthenticate struct {
	*config.AppCfg
}

func (auth AuthAuthenticate) AuthAdmin(c *fiber.Ctx) error {
	return c.Next()
}

func (auth AuthAuthenticate) Auth() fiber.Handler {
	return config.Authenticate(
		auth.AppCfg,
		config.AuthConfig{
			SigningKey: []byte(auth.JwtSecrets.Secret),
		})
}

type authMiddleWare struct {
	authType IAuthenticate
}

func (middleware *authMiddleWare) AuthenticateAdmin(c *fiber.Ctx) error {
	return middleware.authType.AuthAdmin(c)
}

func (middleware *authMiddleWare) Authenticate() fiber.Handler {
	return middleware.authType.Auth()
}

// NewAuthMiddleWare returns new AuthMiddleWare.
func NewAuthMiddleWare(authType IAuthenticate) IAuthenticateMiddleWare {
	return &authMiddleWare{
		authType: authType,
	}
}
