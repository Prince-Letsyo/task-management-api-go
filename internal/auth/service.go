package auth

import (
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/types"
	"github.com/pkg/errors"
)

type RegisterServiceConfiguration func(rs *RegisterService) error

type RegisterService struct {
	IRegisterRepository
}

// NewRegisterService returns new RegisterService.
func newRegisterService(cfgs ...RegisterServiceConfiguration) (*RegisterService, error) {
	rs := &RegisterService{}

	for _, cfg := range cfgs {
		if err := cfg(rs); err != nil {
			return nil, err
		}
	}
	return rs, nil
}

func withDatabaseRegisterRepository(userService types.IUserService, appConfig *config.AppCfg) RegisterServiceConfiguration {
	return withRegisterRepository(NewDBRegisterRepository(userService, appConfig))
}

func withRegisterRepository(registerServiceRepository IRegisterRepository) RegisterServiceConfiguration {
	return func(registerService *RegisterService) error {
		registerService.IRegisterRepository = registerServiceRepository
		return nil
	}
}

func (registerService *RegisterService) NewRegisterService(register *types.RegisterForm) (*types.RegisterForm, error) {
	r, errbook := registerService.store(register)
	if errbook != nil {
		return nil, errbook
	}
	return r, nil
}

func (registerService *RegisterService) ModifyPassword(id uint, passwordReset *types.PasswordResetForm) (*types.RegisterForm, error) {
	b, err := registerService.updatePassword(id, passwordReset)
	if err != nil {
		return nil, errors.Wrap(err, "cannot update user password")
	}
	return b, nil
}

type LoginServiceConfiguration func(ls *LoginService) error

type LoginService struct {
	types.IUserService
	*config.AppCfg
}

// NewLoginService returns new LoginService.
func newLoginService(cfgs ...LoginServiceConfiguration) (*LoginService, error) {
	ls := &LoginService{}

	for _, cfg := range cfgs {
		if err := cfg(ls); err != nil {
			return nil, err
		}
	}
	return ls, nil
}

func withDatabaseLoginRepository(userService types.IUserService) LoginServiceConfiguration {
	return withLoginRepository(userService)
}

func withLoginRepository(userService types.IUserService) LoginServiceConfiguration {
	return func(loginService *LoginService) error {
		loginService.IUserService = userService
		return nil
	}
}

func (loginService *LoginService) CheckLogin(login *types.Login) (*types.User, error) {
	user := &types.User{}
	user, errUser := loginService.ViewByEmail(login.Email, user)

	if errUser != nil {
		return nil, errUser
	} else if !user.EmailVerified {
		return user, errors.New(`Your account is not verified check your mail for verification link`)
	}
	match, _ := loginService.AppCfg.Hash.Match(login.Password, user.Password)
	if !match {
		return nil, errors.New("invalid Username or Password")
	}
	return user, nil
}
