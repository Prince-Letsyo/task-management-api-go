package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID            uint `gorm:"primarykey"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	FirstName     string         `json:"first_name" gorm:"first_name"` //nolint:gofmt
	LastName      string         `json:"last_name" gorm:"last_name"`   //nolint:gofmt
	Email         string         `json:"email" gorm:"email"`
	Password      string         `json:"-" gorm:"password"`
	EmailVerified bool           `json:"email_verified" gorm:"email_verified"`
}
