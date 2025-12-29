package middleware

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Prince-Letsyo/task-management-api-go/cmd/app"
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
		CheckPasswordResetEmailToken(c *fiber.Ctx) error
		CheckPasswordResetToken(c *fiber.Ctx) error
		CheckPasswordResetData(c *fiber.Ctx, passwordReset *types.PasswordResetForm) error
	}
)

type passwordResetMiddleWare struct {
	passwordResetType IPasswordResetValidator
}

type PasswordResetMiddleWare struct {
	types.IUserService
	types.IRegisterService
}

func (passwordResetMiddleWare PasswordResetMiddleWare) CheckPasswordResetEmailToken(c *fiber.Ctx) error {
	token := c.Query("t")
	err := _validatePasswordReset(c, token)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error":   err.Error(),
				"message": "Invalid token",
			})
	}
	return c.Next()
}

func (passwordResetMiddleWare PasswordResetMiddleWare) CheckPasswordResetToken(c *fiber.Ctx) error {
	token := &types.PasswordResetTokenForm{}
	if err := pkg.CustomBodyParser(c, token); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Validation error",
		})
	}
	err := _validatePasswordReset(c, token.Token)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Validation error",
		})
	}
	c.Locals("token_key", token)
	return c.Next()
}

func (passwordResetMiddleWare PasswordResetMiddleWare) CheckPasswordResetData(c *fiber.Ctx, passwordReset *types.PasswordResetForm) error {
	if err := pkg.CustomBodyParser(c, passwordReset); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Validation error",
		})
	}
	if err := _validatePasswordReset(c, passwordReset.Token); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Validation error",
		})
	}
	email := c.Locals("email").(string)
	user := &types.User{}
	_, err := passwordResetMiddleWare.IUserService.ViewByEmail(email, user)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Invalid email address",
		})
	}
	_, errPasswordReset := passwordResetMiddleWare.IRegisterService.ModifyPassword(user.ID, passwordReset)

	if errPasswordReset != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   errPasswordReset.Error(),
			"message": "Error on request password reset request",
		})
	}
	return c.Next()
}

func (middleWare *passwordResetMiddleWare) ValidatePasswordReset(c *fiber.Ctx) error {
	return middleWare.passwordResetType.CheckPasswordResetEmailToken(c)
}

func (middleWare *passwordResetMiddleWare) ValidatePasswordResetToken(c *fiber.Ctx) error {
	return middleWare.passwordResetType.CheckPasswordResetToken(c)
}

func (middleWare *passwordResetMiddleWare) ValidatePasswordResetData(c *fiber.Ctx) error {
	passwordReset := &types.PasswordResetForm{}
	return middleWare.passwordResetType.CheckPasswordResetData(c, passwordReset)
}

func _validatePasswordReset(c *fiber.Ctx, t string) error {
	t = pkg.Decrypt(t, app.Http.Server.Key)
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

func NewPasswordResetMiddleWare(passwordResetType IPasswordResetValidator) IPasswordResetMiddleWare {
	return &passwordResetMiddleWare{
		passwordResetType: passwordResetType,
	}
}
