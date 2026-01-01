package middleware

import (
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/service"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/types"
	"github.com/gofiber/fiber/v2"
)

type (
	ILoginMiddleWare interface {
		ValidateLoginPost(c *fiber.Ctx) error
		RedirectToHomePageOnLogin(c *fiber.Ctx) error
	}

	ILoginValidator interface {
		Validate(c *fiber.Ctx, login *types.Login) error
		RedirectLogin(c *fiber.Ctx) error
	}
)

type loginMiddleWare struct {
	loginType ILoginValidator
}

func (middleWare *loginMiddleWare) ValidateLoginPost(c *fiber.Ctx) error {
	login := &types.Login{}
	return middleWare.loginType.Validate(c, login)
}

func (middleWare *loginMiddleWare) RedirectToHomePageOnLogin(c *fiber.Ctx) error {
	return middleWare.loginType.RedirectLogin(c)
}

type LoginMiddleWare struct {
	types.IUserService
	types.ILoginService
	config.IJWT
}

func (loginMiddleWare LoginMiddleWare) RedirectLogin(c *fiber.Ctx) error {
	redirectService := service.NewAccountAdapterService(
		service.AccountAdapterService{
			IUserService: loginMiddleWare.IUserService,
			IJWT:         loginMiddleWare.IJWT,
		},
	)
	user, err := redirectService.User(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"error":   "Unauthorized",
				"message": err.Error(),
			})
	}
	c.Locals("user", user)

	return c.Next()
}

func (loginMiddleWare LoginMiddleWare) Validate(c *fiber.Ctx, login *types.Login) error {
	if err := pkg.CustomBodyParser(c, login); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Error on validation",
		})
	}

	user, err := loginMiddleWare.CheckLogin(login)
	if user != nil && err != nil {
		c.Locals("err", err.Error())
	} else if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Db Error",
		})
	}
	c.Locals("user", user)
	return c.Next()
}

func NewLoginMiddleWare(loginType ILoginValidator) ILoginMiddleWare {
	return &loginMiddleWare{
		loginType: loginType,
	}
}
