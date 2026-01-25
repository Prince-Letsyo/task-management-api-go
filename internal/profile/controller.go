// Package profile for user profile
package profile

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/middleware"
	"github.com/Prince-Letsyo/task-management-api-go/internal/service"
	"github.com/Prince-Letsyo/task-management-api-go/internal/types"
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
}

type profileController struct {
	Profile
	*config.AppCfg
}

func (controller *profileController) getProfile() fiber.Handler {
	return func(context *fiber.Ctx) error {
		data := fiber.Map{}
		userID, err := service.NewAccountService(
			controller.AppCfg,
			controller.userService,
		).UserID(context)
		if err != nil {
			data["error"] = "Unauthorized user"
			return context.Status(fiber.StatusUnauthorized).JSON(data)
		}
		userProfile := &types.Profile{}
		userProfile, err = controller.profileService.View(userID, userProfile)
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
		userID, err := service.NewAccountService(
			controller.AppCfg,
			controller.userService,
		).UserID(context)
		if err != nil {
			data["error"] = "Unauthorized user"
			return context.Status(fiber.StatusUnauthorized).JSON(data)
		}

		userProfile := &types.Profile{}
		userProfile, err = controller.profileService.Modify(userID, userProfile)
		if err != nil {
			data["error"] = err.Error()
			return context.Status(fiber.StatusBadRequest).JSON(data)
		}
		data["profile"] = userProfile
		return context.Status(fiber.StatusOK).JSON(data)
	}
}

func newProfileController(
	userProfile Profile,
	appCfg *config.AppCfg,
) IProfileController {
	controller := &profileController{
		Profile: userProfile,
		AppCfg:  appCfg,
	}
	authenticateMiddleware := middleware.NewAuthMiddleWare(middleware.AuthAuthenticate{
		AppCfg: appCfg,
	})
	controller.router.Get(
		"",
		authenticateMiddleware.Authenticate(),
		controller.getProfile())
	controller.router.Put(
		"update",
		authenticateMiddleware.Authenticate(),
		controller.updateProfile())

	return controller
}
