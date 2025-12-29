package auth

import (
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/user"
	"github.com/gofiber/fiber/v2"
)

func LoadAuthRoutes(router fiber.Router) {
	userService, err := user.NewUserService(user.WithDatabaseUserRepository())
	if err != nil {
		panic(err.Error())
	}
	registerService, err := newRegisterService(withDatabaseRegisterRepository())
	if err != nil {
		panic(err.Error())
	}

	loginService, err := newLoginService(withDatabaseLoginRepository())
	if err != nil {
		panic(err.Error())
	}

	if _, err := newHttpController(
		withAuthRepository(
			Auth{
				router:          router.Group("/auth"),
				userService:     userService,
				registerService: registerService,
				loginService:    loginService,
			}, config.NewJWT())); err != nil {
		panic(err.Error())
	}
}
