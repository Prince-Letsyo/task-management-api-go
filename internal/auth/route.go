// Package auth implements the authentication routes for the auth.
package auth

import (
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/user"
	"github.com/gofiber/fiber/v2"
)

func LoadAuthRoutes(router fiber.Router, appCfg *config.AppCfg) {
	userService, err := user.NewUserService(
		appCfg,
		user.WithDatabaseUserRepository())
	if err != nil {
		panic(err.Error())
	}
	registerService, err := newRegisterService(
		withDatabaseRegisterRepository(userService))
	if err != nil {
		panic(err.Error())
	}

	loginService, err := newLoginService(
		appCfg,
		withDatabaseLoginRepository(userService))
	if err != nil {
		panic(err.Error())
	}

	if _, err := newHTTPController(
		withAuthController(
			Auth{
				router:          router.Group("/auth/"),
				userService:     userService,
				registerService: registerService,
				loginService:    loginService,
			}, appCfg)); err != nil {
		panic(err.Error())
	}
}
