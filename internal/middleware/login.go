// Package middleware
package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/service"
	"github.com/Prince-Letsyo/task-management-api-go/internal/types"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
)

type (
	ILoginMiddleWare interface {
		ValidateLoginPost(c *fiber.Ctx) error
		RedirectToHomePageOnLogin(c *fiber.Ctx) error
	}

	ILoginValidator interface {
		Validate(c *fiber.Ctx, login *types.Login) error
		RedirectLogin(c *fiber.Ctx, appCfg *config.AppCfg) error
	}
)

type loginMiddleWare struct {
	loginType ILoginValidator
	*config.AppCfg
}

func (middleWare *loginMiddleWare) ValidateLoginPost(c *fiber.Ctx) error {
	login := &types.Login{}
	return middleWare.loginType.Validate(c, login)
}

func (middleWare *loginMiddleWare) RedirectToHomePageOnLogin(c *fiber.Ctx) error {
	return middleWare.loginType.RedirectLogin(c, middleWare.AppCfg)
}

type LoginMiddleWare struct {
	types.IUserService
	types.ILoginService
}

func (loginMiddleWare LoginMiddleWare) RedirectLogin(c *fiber.Ctx, appCfg *config.AppCfg) error {
	fmt.Println("RedirectLogin")
	redirectService := service.NewAccountService(
		appCfg,
		loginMiddleWare.IUserService,
	)
	fmt.Printf("redirectService: %v", redirectService)
	user, err := redirectService.User(c)
	fmt.Printf("RedirectToHomePageOnLogin::user: %v", user)
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
	if err := c.BodyParser(login); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "failed parsinng request body",
		})
	}
	if errs := pkg.ValidateStruct(login); len(errs) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": errs,
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

func NewLoginMiddleWare(loginType ILoginValidator, appCfg *config.AppCfg) ILoginMiddleWare {
	return &loginMiddleWare{
		loginType: loginType,
		AppCfg:    appCfg,
	}
}
