package auth

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AccessClaimsFromRequestWithSession(c *fiber.Ctx, service IAuthService) (*AuthClaims, error) {
	authHeader := c.Get(fiber.HeaderAuthorization)
	if authHeader == "" {
		return nil, errors.New("missing authorization header")
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, errors.New("invalid authorization header")
	}
	return service.ValidateAccessToken(parts[1])
}
