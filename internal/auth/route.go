package auth

import (
	"github.com/Prince-Letsyo/task-management-api-go/cmd/app"
	"github.com/Prince-Letsyo/task-management-api-go/internal/user"
	"github.com/gofiber/fiber/v2"
)

func LoadAuthRoutes(router fiber.Router) {
	userService, err := user.NewUserService(user.WithDatabaseUserRepository(app.HTTP))
	if err != nil {
		panic(err.Error())
	}
	registerService, err := newRegisterService(
		withDatabaseRegisterRepository(userService, app.HTTP))
	if err != nil {
		panic(err.Error())
	}

	loginService, err := newLoginService(
		withDatabaseLoginRepository(userService))
	if err != nil {
		panic(err.Error())
	}

	if _, err := newHTTPController(
		withAuthRepository(
			Auth{
				router:          router.Group("/auth/"),
				userService:     userService,
				registerService: registerService,
				loginService:    loginService,
			}, app.HTTP)); err != nil {
		panic(err.Error())
	}
}
