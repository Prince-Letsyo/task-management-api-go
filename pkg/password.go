package pkg

import (
	"github.com/alexedwards/argon2id"
)

type Password struct {
	// Argon2id configuration
	Params *argon2id.Params
}

func (d *Password) Create(password string) (hash string, err error) {
	return argon2id.CreateHash(password, d.Params)
}

func (d *Password) Match(password string, hash string) (match bool, err error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

func NewPassword() *Password {
	return &Password{
		Params: argon2id.DefaultParams,
	}
}

var DefaultPassword = NewPassword()
