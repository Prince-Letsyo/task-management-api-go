package user

import (
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/types"
	"github.com/pkg/errors"
)

type UserService struct {
	IUserRepository
}

func (us *UserService) ViewByID(id uint, user *types.User) (*types.User, error) {
	u, errUser := us.retrieveByID(id, user)
	if errUser != nil {
		return nil, errUser
	}
	return u, nil
}

func (us *UserService) ViewByEmail(email string, user *types.User) (*types.User, error) {
	u, errUser := us.retrieveByEmail(email, user)
	if errUser != nil {
		return nil, errUser
	}
	return u, nil
}

func (us *UserService) ViewVerifiedUserByEmail(email string, user *types.User) (*types.User, error) {
	u, errUser := us.retrieveVerifiedUserByEmail(email, user)
	if errUser != nil {
		return nil, errUser
	}
	return u, nil
}

func (us *UserService) List(filters *types.UserFilters) (*types.UserPage, error) {
	p, err := us.retrievePage(filters)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (us *UserService) Modify(id uint, user *types.User) (*types.User, error) {
	b, err := us.update(id, user)
	if err != nil {
		return nil, errors.Wrap(err, "cannot update user")
	}
	return b, nil
}

type UserServiceConfiguration func(us *UserService) error

// NewUserService returns new UserService.
func NewUserService(cfgs ...UserServiceConfiguration) (*UserService, error) {
	us := &UserService{}

	for _, cfg := range cfgs {
		if err := cfg(us); err != nil {
			return nil, err
		}
	}
	return us, nil
}

func WithDatabaseUserRepository(appConfig *config.AppCfg) UserServiceConfiguration {
	userRepository := NewDBUser(appConfig)
	return withUserRepository(userRepository)
}

func withUserRepository(userRepository IUserRepository) UserServiceConfiguration {
	return func(userService *UserService) error {
		userService.IUserRepository = userRepository
		return nil
	}
}
