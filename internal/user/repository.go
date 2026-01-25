// Package user hjjk
package user

import (
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
	"github.com/Prince-Letsyo/task-management-api-go/internal/model"
	"github.com/Prince-Letsyo/task-management-api-go/internal/types"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type IUserRepository interface {
	store(user *types.User) (*types.User, error)
	retrieveByID(id uint, user *types.User) (*types.User, error)
	retrieveByEmail(email string, user *types.User) (*types.User, error)
	retrieveVerifiedUserByEmail(email string, user *types.User) (*types.User, error)
	retrievePage(filters *types.UserFilters) (*types.UserPage, error)
	update(id uint, updatedValues map[string]interface{}) (*types.User, error)
}

type DBUserRepository struct {
	*config.AppCfg
}

func (dbUserRepository *DBUserRepository) typeModel(user *types.User) *model.User {
	var profile *model.Profile
	if user.Profile != nil {
		profile = &model.Profile{
			Model: gorm.Model{
				ID: user.Profile.ID,
			},
			Bio:       user.Profile.Bio,
			AvatarURL: user.Profile.AvatarURL,
		}
	}
	return &model.User{
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Email:        user.Email,
		Password:     user.Password,
		IsVerified:   user.IsVerified,
		IsAdmin:      user.IsAdmin,
		UserName:     user.UserName,
		TOTPSecret:   user.TOTPSecret,
		Is2FAEnabled: user.Is2FAEnabled,
		Profile:      profile,
	}
}

func (dbUserRepository *DBUserRepository) modelType(userModel *model.User) *types.User {
	if userModel == nil {
		return nil
	}

	var profile *types.Profile
	if userModel.Profile != nil {
		profile = &types.Profile{
			ID:        userModel.Profile.ID,
			Bio:       userModel.Profile.Bio,
			AvatarURL: userModel.Profile.AvatarURL,
			UserID:    userModel.Profile.UserID,
		}
	}

	return &types.User{
		ID:           userModel.ID,
		FirstName:    userModel.FirstName,
		LastName:     userModel.LastName,
		Password:     userModel.Password,
		Email:        userModel.Email,
		IsVerified:   userModel.IsVerified,
		IsAdmin:      userModel.IsAdmin,
		UserName:     userModel.UserName,
		TOTPSecret:   userModel.TOTPSecret,
		Is2FAEnabled: userModel.Is2FAEnabled,
		Profile:      profile,
	}
}

func (dbUserRepository *DBUserRepository) store(user *types.User) (*types.User, error) {
	userModel := dbUserRepository.typeModel(user)
	if err := dbUserRepository.Database.Create(userModel); err.Error != nil {
		return nil, err.Error
	}
	return dbUserRepository.modelType(userModel), nil
}

func (dbUserRepository *DBUserRepository) retrieveByID(id uint, user *types.User) (*types.User, error) {
	var userModel *model.User
	if err := dbUserRepository.Database.
		Preload("Profile").First(userModel, id); err.Error != nil {
		if errors.Is(err.Error, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, err.Error
	}
	converted := dbUserRepository.modelType(userModel)
	if user != nil {
		*user = *converted
	}
	return user, nil
}

func (dbUserRepository *DBUserRepository) retrieveByEmail(email string, user *types.User) (*types.User, error) {
	userModel := &model.User{}
	if err := dbUserRepository.Database.Where(&model.User{Email: email}).
		Preload("Profile").
		First(userModel); err.Error != nil {

		if errors.Is(err.Error, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, errors.Wrap(err.Error, "failed to retrieve user by email")
	}
	converted := dbUserRepository.modelType(userModel)
	if user != nil {
		*user = *converted
	}
	return user, nil
}

func (dbUserRepository *DBUserRepository) retrieveVerifiedUserByEmail(email string, user *types.User) (*types.User, error) {
	userModel := &model.User{}
	if err := dbUserRepository.Database.
		Where(&model.User{Email: email, IsVerified: true}).
		Preload("Profile").
		First(userModel); err.Error != nil {
		if errors.Is(err.Error, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, errors.Wrap(err.Error, "failed to retrieve verified user by email")
	}
	converted := dbUserRepository.modelType(userModel)
	if user != nil {
		*user = *converted
	}
	return user, nil
}

func (dbUserRepository *DBUserRepository) retrievePage(filters *types.UserFilters) (*types.UserPage, error) {
	userModels := []model.User{}

	if err := dbUserRepository.Database.
		Preload("Profile").
		Model(&model.User{}).
		Scopes(dbUserRepository.Database.Paginate(filters.Filters)).
		Find(userModels); err.Error != nil {
		return nil, err.Error
	}
	users := make([]types.User, len(userModels))
	for i, userModel := range userModels {
		users[i] = *dbUserRepository.modelType(&userModel)
	}

	return &types.UserPage{
		Page: pkg.Page{
			Filters: pkg.Filters{
				Limit:  filters.Limit,
				Offset: filters.Offset,
				Total:  filters.Total,
			},
			Metadata: map[string]any{"description": "filtered page of users"},
		},
		Data: &users,
	}, nil
}

func (dbUserRepository *DBUserRepository) update(id uint, updatedValues map[string]interface{}) (user *types.User, err error) {
	userModel := &model.User{}
	if err := dbUserRepository.Database.Model(&model.User{}).
		Preload("Profile").
		Where("id = ? ", id).
		Updates(updatedValues).First(userModel); err.Error != nil {

		if errors.Is(err.Error, gorm.ErrInvalidDB) {
			return nil, pkg.ErrBookNotFound
		}
		return nil, errors.Wrap(err.Error, "cannot update user")
	}
	user = dbUserRepository.modelType(userModel)

	return user, nil
}

func NewDBUser(appCfg *config.AppCfg) *DBUserRepository {
	return &DBUserRepository{
		AppCfg: appCfg,
	}
}
