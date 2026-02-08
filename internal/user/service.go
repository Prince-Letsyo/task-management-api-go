package user

import (
	"github.com/oarkflow/log"

	"github.com/Prince-Letsyo/task-management-api-go/internal/types"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
)

type UserService struct {
	repo IUserRepository
}

func (us *UserService) Save(user *types.User) (*types.User, error) {
	u, err := us.repo.store(user)
	if err != nil {
		log.Error().Err(err).Str("email", user.Email).Msg("failed to save user")
		return nil, pkg.NewAppError(500, "cannot create user")
	}
	log.Info().Uint("user_id", u.ID).Msg("user saved successfully")
	return u, nil
}

func (us *UserService) ViewByID(id uint, user *types.User) (*types.User, error) {
	u, errUser := us.repo.retrieveByID(id, user)
	if errUser != nil {
		return nil, errUser
	}
	return u, nil
}

func (us *UserService) ViewByEmail(email string, user *types.User) (*types.User, error) {
	u, errUser := us.repo.retrieveByEmail(email, user)
	if errUser != nil {
		return nil, errUser
	}
	return u, nil
}

func (us *UserService) ViewVerifiedUserByEmail(email string, user *types.User) (*types.User, error) {
	u, errUser := us.repo.retrieveVerifiedUserByEmail(email, user)
	if errUser != nil {
		return nil, errUser
	}
	return u, nil
}

func (us *UserService) List(filters *types.UserFilters) (*types.UserPage, error) {
	p, err := us.repo.retrievePage(filters)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (us *UserService) Modify(id uint, updatedValues map[string]interface{}) (*types.User, error) {
	b, err := us.repo.update(id, updatedValues)
	if err != nil {
		log.Error().Err(err).Uint("user_id", id).Msg("failed to update user")
		return nil, pkg.NewAppError(500, "cannot update user")
	}
	log.Info().Uint("user_id", b.ID).Msg("user updated successfully")
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

func WithDatabaseUserRepository(repo IUserRepository) UserServiceConfiguration {
	return func(us *UserService) error {
		us.repo = repo
		return nil
	}
}
