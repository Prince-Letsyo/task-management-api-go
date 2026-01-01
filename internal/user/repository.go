// Package user hjjk
package user

import (
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/model"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/types"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type IUserRepository interface {
	retrieveByID(id uint, user *types.User) (*types.User, error)
	retrieveByEmail(email string, user *types.User) (*types.User, error)
	retrieveVerifiedUserByEmail(email string, user *types.User) (*types.User, error)
	retrievePage(filters *types.UserFilters) (*types.UserPage, error)
	update(id uint, user *types.User) (*types.User, error)
}

type DBUserRepository struct {
	*config.AppCfg
}

func (dbUserRepository *DBUserRepository) retrieveByID(id uint, user *types.User) (*types.User, error) {
	connDB := dbUserRepository.Database.Begin()
	if err := connDB.Model(&model.User{}).
		Where("id = ? ", id).First(user); err.Error != nil {
		defer connDB.Rollback()

		if errors.Is(err.Error, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, err.Error
	}
	return user, nil
}

func (dbUserRepository *DBUserRepository) retrieveByEmail(email string, user *types.User) (*types.User, error) {
	connDB := dbUserRepository.Database.Begin()
	if err := connDB.Where(&model.User{Email: email}).
		First(&user); err.Error != nil {
		defer connDB.Rollback()

		if errors.Is(err.Error, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, err.Error
	}
	return user, nil
}

func (dbUserRepository *DBUserRepository) retrieveVerifiedUserByEmail(email string, user *types.User) (*types.User, error) {
	connDB := dbUserRepository.Database.Begin()
	if err := connDB.
		Where(&model.User{Email: email, EmailVerified: true}).
		First(&user); err.Error != nil {
		defer connDB.Rollback()
		if errors.Is(err.Error, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, err.Error
	}
	return user, nil
}

func (dbUserRepository *DBUserRepository) retrievePage(filters *types.UserFilters) (*types.UserPage, error) {
	users := []types.User{}
	err := dbUserRepository.Database.Begin().
		Model(&model.User{}).
		Scopes(dbUserRepository.Database.Paginate(&filters.Filters)).
		Find(&users)
	if err.Error != nil {
		return nil, err.Error
	}
	return &types.UserPage{
		Page: pkg.Page{
			Filters: pkg.Filters{
				Limit:  filters.Limit,
				Offset: filters.Offset,
				Total:  filters.Total,
			},
			Metadata: map[string]interface{}{"description": "filtered page of users"},
		},
		Data: &users,
	}, nil
}

func (dbUserRepository *DBUserRepository) update(id uint, user *types.User) (*types.User, error) {
	connDB := dbUserRepository.Database.Begin()
	defer connDB.Commit()
	if err := connDB.Model(&model.User{}).
		Where("id = ? ", id).
		Updates(user); err.Error != nil {

		defer connDB.Rollback()
		if errors.Is(err.Error, gorm.ErrInvalidDB) {
			return nil, pkg.ErrBookNotFound
		}
		return nil, err.Error
	}
	return user, nil
}

func NewDBUser(appConfig *config.AppCfg) *DBUserRepository {
	return &DBUserRepository{
		AppCfg: appConfig,
	}
}
