package middleware

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/types"
	"github.com/gofiber/fiber/v2"
)

type (
	IPasswordResetMiddleWare interface {
		ValidatePasswordReset(c *fiber.Ctx) error
		ValidatePasswordResetToken(c *fiber.Ctx) error
		ValidatePasswordResetData(c *fiber.Ctx) error
	}
	IPasswordResetValidator interface {
		CheckPasswordResetEmailToken(c *fiber.Ctx, appCfg *config.AppCfg) error
		CheckPasswordResetToken(c *fiber.Ctx, appCfg *config.AppCfg) error
		CheckPasswordResetData(c *fiber.Ctx, appCfg *config.AppCfg, passwordReset *types.PasswordResetForm) error
	}
)

type passwordResetMiddleWare struct {
	passwordResetType IPasswordResetValidator
	*config.AppCfg
}

type PasswordResetMiddleWare struct {
	types.IUserService
	types.IRegisterService
}

func (passwordResetMiddleWare PasswordResetMiddleWare) CheckPasswordResetEmailToken(c *fiber.Ctx, appCfg *config.AppCfg) error {
	token := c.Query("t")
	err := _validatePasswordReset(c, appCfg, token)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error":   err.Error(),
				"message": "Invalid token",
			})
	}
	return c.Next()
}

func (passwordResetMiddleWare PasswordResetMiddleWare) CheckPasswordResetToken(c *fiber.Ctx, appCfg *config.AppCfg) error {
	token := &types.PasswordResetTokenForm{}
	if err := c.BodyParser(token); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "failed parsing request body",
		})
	}
	if errs := pkg.ValidateStruct(token); len(errs) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": errs,
		})
	}
	err := _validatePasswordReset(c, appCfg, token.Token)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Validation error",
		})
	}
	c.Locals("token_key", token)
	return c.Next()
}

func (passwordResetMiddleWare PasswordResetMiddleWare) CheckPasswordResetData(c *fiber.Ctx, appCfg *config.AppCfg, passwordReset *types.PasswordResetForm) error {
	if err := c.BodyParser(passwordReset); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "failed parsing request body",
		})
	}
	if errs := pkg.ValidateStruct(passwordReset); len(errs) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": errs,
		})
	}
	if err := _validatePasswordReset(c, appCfg, passwordReset.Token); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Validation error",
		})
	}
	email := c.Locals("email").(string)
	user := &types.User{}
	_, err := passwordResetMiddleWare.ViewByEmail(email, user)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Invalid email address",
		})
	}
	_, errPasswordReset := passwordResetMiddleWare.ModifyUserPassword(user.ID, passwordReset)

	if errPasswordReset != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   errPasswordReset.Error(),
			"message": "Error on request password reset request",
		})
	}
	return c.Next()
}

func (middleWare *passwordResetMiddleWare) ValidatePasswordReset(c *fiber.Ctx) error {
	return middleWare.passwordResetType.CheckPasswordResetEmailToken(c, middleWare.AppCfg)
}

func (middleWare *passwordResetMiddleWare) ValidatePasswordResetToken(c *fiber.Ctx) error {
	return middleWare.passwordResetType.CheckPasswordResetToken(c, middleWare.AppCfg)
}

func (middleWare *passwordResetMiddleWare) ValidatePasswordResetData(c *fiber.Ctx) error {
	passwordReset := &types.PasswordResetForm{}
	return middleWare.passwordResetType.CheckPasswordResetData(c, middleWare.AppCfg, passwordReset)
}

func _validatePasswordReset(c *fiber.Ctx, appCfg *config.AppCfg, t string) error {
	t, err := pkg.Decrypt(t, appCfg.Server.Key)
	if err != nil {
		return errors.New("invalid password reset token")
	}
	emailParts := strings.Split(t, "-reset-")

	if len(emailParts) != 2 {
		return errors.New("invalid password reset token")
	}

	tokenTS, err := strconv.ParseInt(emailParts[1], 10, 64)
	if err != nil {
		return errors.New("invalid password reset token")
	}
	now := time.Now().Unix()
	diff := now - tokenTS
	if diff > (5 * 60) {
		return errors.New("password reset token has expired")
	} else if diff < 0 {
		return errors.New("invalid password reset token")
	}
	c.Locals("email", emailParts[0])
	c.Locals("token", t)
	return nil
}

func NewPasswordResetMiddleWare(passwordResetType IPasswordResetValidator, appCfg *config.AppCfg) IPasswordResetMiddleWare {
	return &passwordResetMiddleWare{
		passwordResetType: passwordResetType,
		AppCfg:            appCfg,
	}
}
