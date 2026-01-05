package user

import (
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/types"
	"github.com/pkg/errors"
)

type UserService struct {
	IUserRepository
	*config.AppCfg
}

func (us *UserService) Save(user *types.User) (*types.User, error) {
	u, err := us.store(user)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create user")
	}
	return u, nil
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
func NewUserService(appCfg *config.AppCfg, cfgs ...UserServiceConfiguration) (*UserService, error) {
	us := &UserService{
		AppCfg: appCfg,
	}

	for _, cfg := range cfgs {
		if err := cfg(us); err != nil {
			return nil, err
		}
	}
	return us, nil
}

func WithDatabaseUserRepository() UserServiceConfiguration {
	return func(us *UserService) error {
		us.IUserRepository = NewDBUser(us.AppCfg)
		return nil
	}
}
