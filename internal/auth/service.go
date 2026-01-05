package auth

import (
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/types"
	"github.com/pkg/errors"
)

type RegisterServiceConfiguration func(rs *RegisterService) error

type RegisterService struct {
	types.IUserService
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

func withDatabaseRegisterRepository(userService types.IUserService) RegisterServiceConfiguration {
	return func(rs *RegisterService) error {
		rs.IUserService = userService
		return nil
	}
}

func (registerService *RegisterService) NewUser(register *types.RegisterForm) (*types.User, error) {
	user := &types.User{
		FirstName: register.FirstName,
		LastName:  register.LastName,
		Email:     register.Email,
		UserName:  register.UserName,
		Password:  register.Password,
	}
	if user, _ = registerService.ViewByEmail(user.Email, user); user != nil {
		return nil, errors.New("user already exists")
	}
	if _, err := registerService.Save(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (registerService *RegisterService) ModifyUserPassword(id uint, passwordReset *types.PasswordResetForm) (*types.User, error) {
	user := &types.User{
		Password: passwordReset.Password,
	}
	if _, err := registerService.Modify(id, user); err != nil {
		return nil, err
	}
	return user, nil
}

type LoginServiceConfiguration func(ls *LoginService) error

type LoginService struct {
	types.IUserService
	*config.AppCfg
}

// NewLoginService returns new LoginService.
func newLoginService(appCfg *config.AppCfg, cfgs ...LoginServiceConfiguration) (*LoginService, error) {
	ls := &LoginService{
		AppCfg: appCfg,
	}

	for _, cfg := range cfgs {
		if err := cfg(ls); err != nil {
			return nil, err
		}
	}
	return ls, nil
}

func withDatabaseLoginRepository(userService types.IUserService) LoginServiceConfiguration {
	return func(ls *LoginService) error {
		ls.IUserService = userService
		return nil
	}
}

func (loginService *LoginService) CheckLogin(login *types.Login) (*types.User, error) {
	user := &types.User{}
	user, errUser := loginService.ViewVerifiedUserByEmail(login.Email, user)

	if errUser != nil {
		return nil, errUser
	} else if !user.IsVerified {
		return user, errors.New(`Your account is not verified check your mail for verification link`)
	}
	match, _ := pkg.DefaultPassword.Match(login.Password, user.Password)
	if !match {
		return nil, errors.New("invalid Username or Password")
	}
	return user, nil
}
