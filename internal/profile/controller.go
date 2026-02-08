// Package profile for user profile
package profile

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/auth"
	"github.com/Prince-Letsyo/task-management-api-go/internal/types"
	"github.com/Prince-Letsyo/task-management-api-go/internal/user"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
)

type IProfileController interface {
	getProfile() fiber.Handler
	updateProfile() fiber.Handler
}

type Profile struct {
	router         fiber.Router
	profileService types.IProfileService
	userService    types.IUserService
	authService    auth.IAuthService
}

type profileController struct {
	Profile
	*config.AppCfg
}

func (controller *profileController) getProfile() fiber.Handler {
	return func(context *fiber.Ctx) error {
		data := fiber.Map{}
		claims, err := auth.AccessClaimsFromRequestWithSession(context, controller.authService)
		if err != nil {
			data["error"] = "Unauthorized user"
			return context.Status(fiber.StatusUnauthorized).JSON(data)
		}
		userProfile := &types.Profile{}
		userProfile, err = controller.profileService.View(claims.UserID, userProfile)
		if err != nil {
			data["error"] = err.Error()
			return context.Status(fiber.StatusInternalServerError).JSON(data)
		}
		data["profile"] = userProfile
		return context.Status(fiber.StatusOK).JSON(data)
	}
}

func (controller *profileController) updateProfile() fiber.Handler {
	return func(context *fiber.Ctx) error {
		data := fiber.Map{}
		profileBody := &types.ProfileForm{}
		if err := context.BodyParser(profileBody); err != nil {
			data["error"] = "failed parsing request body"
			return context.Status(fiber.StatusUnprocessableEntity).JSON(data)
		}
		if errs := pkg.ValidateStruct(profileBody); len(errs) > 0 {
			data["errors"] = errs
			return context.Status(fiber.StatusBadRequest).JSON(data)
		}
		claims, err := auth.AccessClaimsFromRequestWithSession(context, controller.authService)
		if err != nil {
			data["error"] = "Unauthorized user"
			return context.Status(fiber.StatusUnauthorized).JSON(data)
		}

		userProfile := &types.Profile{}
		userProfile, err = controller.profileService.Modify(claims.UserID, userProfile)
		if err != nil {
			data["error"] = err.Error()
			return context.Status(fiber.StatusBadRequest).JSON(data)
		}
		data["profile"] = userProfile
		return context.Status(fiber.StatusOK).JSON(data)
	}
}

func LoadProfileRoute(router fiber.Router, appCfg *config.AppCfg) {
	userRepo := user.NewDBUser(appCfg.Database.DB)
	userService, err := user.NewUserService(
		user.WithDatabaseUserRepository(userRepo))
	if err != nil {
		panic(err.Error())
	}

	profileRepo := NewDBProfileRepository(appCfg.Database.DB)
	profileService, err := newProfileService(
		withDatabaseProfileRepository(profileRepo))
	if err != nil {
		panic(err.Error())
	}

	authRepo := auth.NewDBAuthRepository(appCfg.Database.DB)
	authService := auth.NewAuthService(authRepo, appCfg.Auth, appCfg.JwtSecrets, appCfg.Server)

	newProfileController(Profile{
		router:         router.Group("/profile/"),
		profileService: profileService,
		userService:    userService,
		authService:    authService,
	}, appCfg)
}

func newProfileController(
	userProfile Profile,
	appCfg *config.AppCfg,
) IProfileController {
	controller := &profileController{
		Profile: userProfile,
		AppCfg:  appCfg,
	}
	controller.router.Get(
		"",
		controller.getProfile())
	controller.router.Put(
		"update",
		controller.updateProfile())

	return controller
}
