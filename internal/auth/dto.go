package auth

import "time"

type UserCreateRequest struct {
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	UserName    string `json:"username" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password_one" validate:"required,min=8"`
	PasswordTwo string `json:"password_two" validate:"required,min=8"`
}

type AuthLoginRequest struct {
	UserName string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

type ActivationEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type TokenRequest struct {
	Token string `json:"token" validate:"required"`
}

type Verify2FARequest struct {
	Token     string `json:"token" validate:"required"`
	TOTPToken string `json:"totp_token" validate:"required"`
}

type PasswordResetRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Token       string `json:"token" validate:"required"`
	PasswordOne string `json:"password_one" validate:"required,min=8"`
	PasswordTwo string `json:"password_two" validate:"required,min=8"`
}

type ChangeEmailRequest struct {
	NewEmail string `json:"new_email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type AccessToken struct {
	Token    string    `json:"token"`
	Duration time.Time `json:"duration"`
}

type RefreshToken struct {
	Token    string    `json:"token"`
	Duration time.Time `json:"duration"`
}

type Temp2FAToken struct {
	Token    string    `json:"token"`
	Duration time.Time `json:"duration"`
}

type TokenModel struct {
	AccessToken  AccessToken  `json:"access_token"`
	RefreshToken RefreshToken `json:"refresh_token"`
}

type UserResponse struct {
	Requires2FA bool        `json:"requires_2fa"`
	Token       interface{} `json:"token"`
}
