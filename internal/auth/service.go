package auth

import (
	"github.com/Prince-Letsyo/task-management-api-go/cmd/app"
	"github.com/Prince-Letsyo/task-management-api-go/internal/user"
	"github.com/pkg/errors"
)

type IRegisterService interface {
	newRegisterServicee(register *RegisterForm) (*RegisterForm, error)
	modifyPassword(id uint, passwordReset *PasswordResetForm) (*RegisterForm, error)
}

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

func (registerService *RegisterService) newRegisterServicee(register *RegisterForm) (*RegisterForm, error) {
	r, errbook := registerService.store(register)
	if errbook != nil {
		return nil, errbook
	}
	return r, nil
}

func (registerService *RegisterService) modifyPassword(id uint, passwordReset *PasswordResetForm) (*RegisterForm, error) {
	b, err := registerService.updatePassword(id, passwordReset)
	if err != nil {
		return nil, errors.Wrap(err, "cannot update user password")
	}
	return b, nil
}

type ILoginService interface {
	checkLogin(login *Login) (*user.User, error)
}

type LoginServiceConfiguration func(ls *LoginService) error

type LoginService struct {
	user.IUserService
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

func withLoginRepository(userService user.IUserService) LoginServiceConfiguration {
	return func(loginService *LoginService) error {
		loginService.IUserService = userService
		return nil
	}
}

func (loginService *LoginService) checkLogin(login *Login) (*user.User, error) {
	user := &user.User{}
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
