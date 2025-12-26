package model

type Login struct {
	Email    string `json:"email" gorm:"email" form:"email" validate:"required|email"`
	Password string `json:"password" gorm:"password" form:"password" validate:"required"`
}

func (Login) TableName() string {
	return "users"
}
