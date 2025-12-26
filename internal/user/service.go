package user

import "github.com/pkg/errors"

type IUserService interface {
	ViewByID(id uint, user *User) (*User, error)
	ViewByEmail(email string, user *User) (*User, error)
	ViewVerifiedUserByEmail(email string, user *User) (*User, error)
	List(filters *UserFilters) (*UserPage, error)
	Modify(id uint, user *User) (*User, error)
}

type UserService struct {
	IUserRepository
}

func (us *UserService) ViewByID(id uint, user *User) (*User, error) {
	u, errUser := us.retrieveById(id, user)
	if errUser != nil {
		return nil, errUser
	}
	return u, nil
}

func (us *UserService) ViewByEmail(email string, user *User) (*User, error) {
	u, errUser := us.retrieveByEmail(email, user)
	if errUser != nil {
		return nil, errUser
	}
	return u, nil
}

func (us *UserService) ViewVerifiedUserByEmail(email string, user *User) (*User, error) {
	u, errUser := us.retrieveVerifiedUserByEmail(email, user)
	if errUser != nil {
		return nil, errUser
	}
	return u, nil
}

func (us *UserService) List(filters *UserFilters) (*UserPage, error) {
	p, err := us.retrievePage(filters)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (us *UserService) Modify(id uint, user *User) (*User, error) {
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

func WithDatabaseUserRepository() UserServiceConfiguration {
	userRepository := NewDBUser()
	return withUserRepository(userRepository)
}

func withUserRepository(userRepository IUserRepository) UserServiceConfiguration {
	return func(userService *UserService) error {
		userService.IUserRepository = userRepository
		return nil
	}
}
