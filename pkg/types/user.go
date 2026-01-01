// Package types jjhg
package types

import "github.com/Prince-Letsyo/task-management-api-go/pkg"

type User struct {
	ID            uint   `json:"id"`
	FirstName     string `json:"first_name" ` //nolint:gofmt
	LastName      string `json:"last_name" `  //nolint:gofmt
	Email         string `json:"email" `
	Password      string `json:"-" `
	EmailVerified bool   `json:"email_verified" `
	IsAdmin       bool   `json:"is_admin" `
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
	ViewByID(id uint, user *User) (*User, error)
	ViewByEmail(email string, user *User) (*User, error)
	ViewVerifiedUserByEmail(email string, user *User) (*User, error)
	List(filters *UserFilters) (*UserPage, error)
	Modify(id uint, user *User) (*User, error)
}
