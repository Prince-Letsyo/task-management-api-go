package core

import (
	appCfg "github.com/Prince-Letsyo/task-management-api-go/cmd/app"
	"github.com/Prince-Letsyo/task-management-api-go/internal/auth"

	"github.com/gofiber/fiber/v2"
)

func LoadRoutes(app *fiber.App) {
	mainRouter := app.Group(appCfg.Http.Server.FullAPIPath())
	auth.LoadAuthRoutes(mainRouter)
	app.Get("*", func(c *fiber.Ctx) error {
		return appCfg.Http.Server.ErrorHandler(c, fiber.NewError(fiber.StatusNotFound, "Page not found"))
	})
}

func setupSwagger(f *fiber.App) {
}
