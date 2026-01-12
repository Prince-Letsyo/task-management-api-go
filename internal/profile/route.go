package profile

import (
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/user"
	"github.com/gofiber/fiber/v2"
)

func LoadProfileRoute(router fiber.Router, appCfg *config.AppCfg) {
	userService, err := user.NewUserService(
		appCfg,
		user.WithDatabaseUserRepository())
	if err != nil {
		panic(err.Error())
	}

	profileService, err := newProfileService(
		appCfg,
		withDatabaseProfileRepository(
			NewDBProfileRepository(appCfg),
		))
	if err != nil {
		panic(err.Error())
	}

	if _, err := newHTTPProfileController(
		withProfileController(Profile{
			router:         router.Group("/profile/"),
			profileService: profileService,
			userService:    userService,
		}, appCfg),
	); err != nil {
		panic(err.Error())
	}
}
