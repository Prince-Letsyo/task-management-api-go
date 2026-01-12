package middleware

import (
	"fmt"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/types"
	"github.com/gofiber/fiber/v2"
)

type (
	IRegisterMiddleWare interface {
		ValidateConfirmToken(c *fiber.Ctx) error
		ValidateRegister(c *fiber.Ctx) error
	}

	IRegisterValidatorToken interface {
		ValidateToken(c *fiber.Ctx, email string, user *types.User, appCfg *config.AppCfg) error
		Validate(c *fiber.Ctx, register *types.RegisterForm) error
	}
)

type registerMiddleWare struct {
	registerType IRegisterValidatorToken
	*config.AppCfg
}

type TokenForm struct {
	T string `json:"t"`
}

func (middleWare *registerMiddleWare) ValidateConfirmToken(c *fiber.Ctx) error {
	user := &types.User{}
	tokenF := &TokenForm{}
	tErr := pkg.CustomQueryParser(c, tokenF)
	if tErr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": tErr.Error()})
	}
	t, err := pkg.Decrypt(tokenF.T, middleWare.Server.Key)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": err.Error()})
	}
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
		return err
	} else {
		fmt.Printf("Error: %v", err)
	}
	r, errRegister := registerMiddleWare.NewUser(register)
	if errRegister != nil {
		return context.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   errRegister.Error(),
			"message": "Error on register request",
		})
	}
	context.Locals("register", r)
	return context.Next()
}

func (registerMiddleWare RegisterMiddleWare) ValidateToken(c *fiber.Ctx, t string, user *types.User, appCfg *config.AppCfg) error {
	_, err := registerMiddleWare.ViewByEmail(t, user)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Error on token verification request",
		})
	}
	if user.IsVerified {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Email was already validated",
			"message": "Error on token verification request",
		})
	}

	vUser, err := registerMiddleWare.Modify(
		user.ID,
		map[string]interface{}{"email_verified": true})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Error on token verification request",
		})
	}

	c.Locals("user", vUser)
	return c.Next()
}

func NewRegisterMiddleWare(registerType IRegisterValidatorToken, appCfg *config.AppCfg) IRegisterMiddleWare {
	return &registerMiddleWare{
		registerType: registerType,
		AppCfg:       appCfg,
	}
}
