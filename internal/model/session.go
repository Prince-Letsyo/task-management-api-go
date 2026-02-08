package model

import "time"

type Session struct {
	ID               string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID           uint       `json:"user_id" gorm:"column:user_id;index;not null"`
	RefreshTokenHash string     `json:"refresh_token_hash" gorm:"column:refresh_token_hash;type:varchar(128);not null"`
	UserAgent        *string    `json:"user_agent" gorm:"column:user_agent;type:varchar(512)"`
	IPAddress        *string    `json:"ip_address" gorm:"column:ip_address;type:varchar(64)"`
	LastUsedAt       *time.Time `json:"last_used_at" gorm:"column:last_used_at"`
	RevokedAt        *time.Time `json:"revoked_at" gorm:"column:revoked_at"`
	CreatedAt        time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"column:updated_at"`
	User             *User      `json:"user" gorm:"constraint:OnDelete:CASCADE;foreignKey:UserID"`
}

func (Session) TableName() string {
	return "session"
}
