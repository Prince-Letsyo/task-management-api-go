package auth

import (
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/user"
	"github.com/gofiber/fiber/v2"
)

func LoadAuthRoutes(router fiber.Router, appConfig *config.AppCfg) {
	userService, err := user.NewUserService(
		appConfig,
		user.WithDatabaseUserRepository())
	if err != nil {
		panic(err.Error())
	}
	registerService, err := newRegisterService(
		appConfig,
		withDatabaseRegisterRepository(userService))
	if err != nil {
		panic(err.Error())
	}

	loginService, err := newLoginService(
		appConfig,
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
			}, appConfig)); err != nil {
		panic(err.Error())
	}
}
