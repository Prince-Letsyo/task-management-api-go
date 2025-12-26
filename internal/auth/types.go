package auth

import (
	"github.com/Prince-Letsyo/task-management-api-go/internal/user"
	"github.com/gofiber/fiber/v2"
)

type RegisterForm struct {
	ID        uint   `json:"id"`
	FirstName string `json:"first_name"  form:"first_name" validate:"required"`
	LastName  string `json:"last_name"  form:"last_name" validate:"required"`
	Email     string `json:"email"  form:"email" validate:"required|email"`
	Password  string `json:"password"  form:"password" validate:"required"`
	CPassword string `json:"c_password" form:"c_password" validate:"required"`
}

type RequestPasswordResetForm struct {
	Email string `json:"email"  form:"email" validate:"required|email"`
}

type Login struct {
	Email    string `json:"email" gorm:"email" form:"email" validate:"required|email"`
	Password string `json:"password" gorm:"password" form:"password" validate:"required"`
}

type PasswordResetTokenForm struct {
	Token string `json:"token"  form:"token" validate:"required"`
}

type PasswordResetForm struct {
	Token     string `json:"token"  validate:"required"`
	Password  string `json:"password"  form:"password" validate:"required"`
	CPassword string `json:"c_password" form:"c_password" validate:"required"`
}

type Auth struct {
	router   fiber.Router
	user     user.IUserService
	register IRegisterService
	login    ILoginService
}
type UserAuth struct {
	router fiber.Router
	user   user.IUserService
}
