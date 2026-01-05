package types

type RegisterForm struct {
	FirstName string `json:"first_name"  form:"first_name" validate:"required"`
	LastName  string `json:"last_name"  form:"last_name" validate:"required"`
	UserName  string `json:"username" form:"username" validate:"required"`
	Email     string `json:"email"  form:"email" validate:"required,email"`
	Password  string `json:"password"  form:"password" validate:"required,min=8"`
	CPassword string `json:"c_password" form:"c_password" validate:"required,eqfield=Password"`
}

type RequestPasswordResetForm struct {
	Email string `json:"email"  form:"email" validate:"required|email"`
}

type Login struct {
	Email    string `json:"email" gorm:"email" form:"email" validate:"required,email"`
	Password string `json:"password" gorm:"password" form:"password" validate:"required,min=8"`
}

type PasswordResetTokenForm struct {
	Token string `json:"token"  form:"token" validate:"required"`
}

type PasswordResetForm struct {
	Token     string `json:"token"  validate:"required"`
	Password  string `json:"password"  form:"password" validate:"required,min=8"`
	CPassword string `json:"c_password" form:"c_password" validate:"required,eqfield=Password"`
}

type IRegisterService interface {
	NewUser(register *RegisterForm) (*User, error)
	ModifyUserPassword(id uint, passwordReset *PasswordResetForm) (*User, error)
}

type ILoginService interface {
	CheckLogin(login *Login) (*User, error)
}
