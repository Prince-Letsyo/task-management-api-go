package auth

import (
	"time"

	"github.com/Prince-Letsyo/task-management-api-go/cmd/app"
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/gofiber/fiber/v2"
)

type authController struct {
	Auth
	IJWT
	config.JwtSecrets
	config time.Duration
}

func (hc *authController) register() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Registered successfully! Please confirm your email",
		})
	}
}

func newAuthController(
	authC Auth,
	authJWT IJWT,
) IAuthController {
	controller := &authController{
		Auth:       authC,
		IJWT:       authJWT,
		JwtSecrets: app.Http.JwtSecrets,
		config:     app.Http.HTTP.Timeout,
	}
	return controller
}
