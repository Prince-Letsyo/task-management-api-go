// Package model jhd
package model

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Prince-Letsyo/task-management-api-go/pkg"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirstName    string   `json:"first_name" gorm:"column:first_name;not null"`
	LastName     string   `json:"last_name" gorm:"column:last_name;not null"`
	UserName     string   `json:"username" gorm:"column:username;not null;uniqueIndex"`
	Email        string   `json:"email" gorm:"column:email;not null;uniqueIndex"`
	Password     string   `json:"-" gorm:"column:password;type:varchar(256);not null"`
	IsVerified   bool     `json:"email_verified" gorm:"column:email_verified;default:false"`
	IsAdmin      bool     `json:"is_admin" gorm:"column:is_admin;default:false"`
	Is2FAEnabled bool     `json:"is_2fa_enabled" gorm:"column:is_2fa_enabled;default:false;not null"`
	TOTPSecret   *string  `json:"totp_secret" gorm:"column:totp_secret;type:varchar(32)"`
	Profile      *Profile `json:"profile" gorm:"constraint:OnDelete:CASCADE;foreignKey:UserID"`
}

func (User) TableName() string {
	return "user"
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	return u.validateAndHashPassword(tx.Statement.Context)
}

func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	if passwordChanged := tx.Statement.Changed("Password"); passwordChanged {
		return u.validateAndHashPassword(tx.Statement.Context)
	}
	return nil
}

func (u *User) AfterCreate(tx *gorm.DB) error {
	if u.Profile != nil && u.Profile.ID != 0 {
		return nil
	}

	return tx.Create(&Profile{
		UserID: u.ID,
	}).Error
}

func (u *User) validateAndHashPassword(ctx context.Context) error {
	plainPassword := u.Password
	if strings.TrimSpace(plainPassword) == "" {
		return errors.New("password cannot be empty")
	}

	result := pkg.DefaultValidator.ValidatePassword(plainPassword, u.UserName, u.Email)

	if !result.IsValid {
		return errors.New(result.Errors[0])
	}

	hashedPassword, err := pkg.DefaultPassword.Create(plainPassword)
	if err != nil {
		return err
	}
	fmt.Printf("hashedPassword: %v\n", hashedPassword)

	u.Password = hashedPassword
	return nil
}
