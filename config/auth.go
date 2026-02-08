package config

import "time"

type AuthSettings struct {
	AdminAPIKey         string        `json:"admin_api_key" yaml:"admin_api_key" env:"ADMIN_API_KEY"`
	TOTPSecretKey       string        `json:"totp_secret_key" yaml:"totp_secret_key" env:"TOTP_SECRET_KEY"`
	Temp2FAExpire       time.Duration `json:"temp_2fa_expire" yaml:"temp_2fa_expire" env:"TEMP_2FA_EXPIRE" env-default:"5m"`
	ActivateExpire      time.Duration `json:"activate_expire" yaml:"activate_expire" env:"ACTIVATE_EXPIRE" env-default:"15m"`
	PasswordResetExpire time.Duration `json:"password_reset_expire" yaml:"password_reset_expire" env:"PASSWORD_RESET_EXPIRE" env-default:"15m"`
}
