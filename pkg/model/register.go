package model

import (
	"gorm.io/gorm"
)

type RegisterForm struct {
	gorm.Model
	FirstName string `json:"first_name" gorm:"first_name" form:"first_name" `
	LastName  string `json:"last_name" gorm:"last_name" form:"last_name" `
	Email     string `json:"email" gorm:"email" form:"email" validate:"required|email"`
	Password  string `json:"password" gorm:"password" form:"password" validate:"required"`
	CPassword string `json:"c_password" form:"c_password" validate:"required" gorm:"-"`
}

func (RegisterForm) TableName() string {
	return "users"
}
