// Package core dfs
package core

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/auth"
	"github.com/Prince-Letsyo/task-management-api-go/internal/profile"
	"github.com/Prince-Letsyo/task-management-api-go/internal/queue"
)

func LoadRoutes(appCfg *config.AppCfg, qClient *queue.WorkerClient) {
	app := appCfg.Server.App
	mainRouter := app.Group(appCfg.Server.FullAPIPath())
	auth.LoadAuthRoutes(mainRouter, appCfg, qClient)
	profile.LoadProfileRoute(mainRouter, appCfg)
	app.Get("*", func(c *fiber.Ctx) error {
		return appCfg.Server.ErrorHandler(c, fiber.NewError(fiber.StatusNotFound, "Page not found"))
	})
}
