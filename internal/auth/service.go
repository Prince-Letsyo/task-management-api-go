package auth

import (
	"github.com/Prince-Letsyo/task-management-api-go/cmd/app"
	"github.com/Prince-Letsyo/task-management-api-go/internal/user"
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

func withDatabaseRegisterRepository() RegisterServiceConfiguration {
	userService, err := user.NewUserService(user.WithDatabaseUserRepository())
	if err != nil {
		panic(err.Error())
	}
	rs := NewDBRegisterRepository(userService)
	return withRegisterRepository(rs)
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

func withDatabaseLoginRepository() LoginServiceConfiguration {
	userService, err := user.NewUserService(user.WithDatabaseUserRepository())
	if err != nil {
		panic(err.Error())
	}
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
	_, errUser := loginService.ViewByEmail(login.Email, user)

	if errUser != nil {
		return nil, errUser
	} else if !user.EmailVerified {
		return user, errors.New(`Your account is not verified check your mail for verification link`)
	}
	match, _ := app.Http.Hash.Match(login.Password, user.Password)
	if !match {
		return nil, errors.New("invalid Username or Password")
	}
	return user, nil
}
