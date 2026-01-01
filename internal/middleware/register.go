package middleware

import (
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/service"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/types"
	"github.com/gofiber/fiber/v2"
)

type (
	IRegisterMiddleWare interface {
		ValidateConfirmToken(c *fiber.Ctx) error
		ValidateRegister(c *fiber.Ctx) error
	}

	IRegisterValidatorToken interface {
		ValidateToken(c *fiber.Ctx, email string, user *types.User, appConfig *config.AppCfg) error
		Validate(c *fiber.Ctx, register *types.RegisterForm) error
	}
)

type registerMiddleWare struct {
	registerType IRegisterValidatorToken
	*config.AppCfg
}

func (middleWare *registerMiddleWare) ValidateConfirmToken(c *fiber.Ctx) error {
	user := &types.User{}
	t := pkg.Decrypt(c.Query("t"), middleWare.AppCfg.Server.Key)
	return middleWare.registerType.ValidateToken(c, t, user, middleWare.AppCfg)
}

func (middleWare *registerMiddleWare) ValidateRegister(c *fiber.Ctx) error {
	register := &types.RegisterForm{}
	return middleWare.registerType.Validate(c, register)
}

type RegisterMiddleWare struct {
	types.IUserService
	types.IRegisterService
}

func (registerMiddleWare RegisterMiddleWare) Validate(context *fiber.Ctx, register *types.RegisterForm) error {
	if err := pkg.CustomBodyParser(context, register); err != nil {
		return context.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error":   err.Error(),
				"message": "Validation Error",
			},
		)
	}
	r, errRegister := registerMiddleWare.NewRegisterService(register)
	if errRegister != nil {
		return context.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   errRegister.Error(),
			"message": "Error on register request",
		})
	}
	context.Locals("register", r)
	return context.Next()
}

func (registerMiddleWare RegisterMiddleWare) ValidateToken(c *fiber.Ctx, t string, user *types.User, appConfig *config.AppCfg) error {
	_, err := registerMiddleWare.ViewByEmail(t, user)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Error on token verification request",
		})
	}

	if user.EmailVerified {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Email was already validated",
			"message": "Error on token verification request",
		})
	}

	user.EmailVerified = true
	vUser, err := registerMiddleWare.Modify(user.ID, user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Error on token verification request",
		})
	}

	errToken := service.
		NewAccountAdapterService(
			service.AccountAdapterService{
				IUserService: registerMiddleWare.IUserService,
				AppCfg:       appConfig,
			},
		).
		Login(c, vUser)
	if errToken != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   errToken.Error(),
			"message": "Error on token verification request",
		})
	}
	c.Locals("user", vUser)
	return c.Next()
}

func NewRegisterMiddleWare(registerType IRegisterValidatorToken, appConfig *config.AppCfg) IRegisterMiddleWare {
	return &registerMiddleWare{
		registerType: registerType,
		AppCfg:       appConfig,
	}
}
