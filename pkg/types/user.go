// Package types jjhg
package types

import "github.com/Prince-Letsyo/task-management-api-go/pkg"

type User struct {
	ID           uint     `json:"id"`
	FirstName    string   `json:"first_name" validate:"required"`
	LastName     string   `json:"last_name" validate:"required"`
	UserName     string   `json:"username" validate:"required"`
	Email        string   `json:"email" validate:"required,email"`
	Password     string   `json:"password" validate:"required, min=8"`
	IsVerified   bool     `json:"email_verified" `
	IsAdmin      bool     `json:"is_admin" `
	Is2FAEnabled bool     `json:"is_2fa_enabled" `
	TOTPSecret   *string  `json:"totp_secret" `
	Profile      *Profile `json:"profile"`
}

type UserFilters struct {
	pkg.Filters
	Name        string `json:"name" query:"name"`
	Description string `json:"description" query:"description"`
}

type UserPage struct {
	pkg.Page
	Data *[]User `json:"data"`
}

type IUserService interface {
	Save(user *User) (*User, error)
	ViewByID(id uint, user *User) (*User, error)
	ViewByEmail(email string, user *User) (*User, error)
	ViewVerifiedUserByEmail(email string, user *User) (*User, error)
	List(filters *UserFilters) (*UserPage, error)
	Modify(id uint, user *User) (*User, error)
}
