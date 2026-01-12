// Package core dfs
package core

import (
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/auth"
	"github.com/Prince-Letsyo/task-management-api-go/internal/profile"

	"github.com/gofiber/fiber/v2"
)

func LoadRoutes(appCfg *config.AppCfg) {
	app := appCfg.Server.App
	mainRouter := app.Group(appCfg.Server.FullAPIPath())
	auth.LoadAuthRoutes(mainRouter, appCfg)
	profile.LoadProfileRoute(mainRouter, appCfg)
	app.Get("*", func(c *fiber.Ctx) error {
		return appCfg.Server.ErrorHandler(c, fiber.NewError(fiber.StatusNotFound, "Page not found"))
	})
}

//func setupSwagger(f *fiber.App) {
//}
