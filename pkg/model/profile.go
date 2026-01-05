package model

import "gorm.io/gorm"

type Profile struct {
	gorm.Model
	UserID    uint    `gorm:"column:user_id;not null;uniqueIndex"`
	Bio       *string `gorm:"column:bio;type:text"`
	AvatarURL *string `gorm:"column:avatar_url;type:text"`
	User      *User   `gorm:"foreignKey:UserID;references:ID"`
}

func (Profile) TableName() string {
	return "profile"
}
