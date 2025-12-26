package user

import (
	"github.com/Prince-Letsyo/task-management-api-go/cmd/app"
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/model"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type IUserRepository interface {
	retrieveById(id uint, user *User) (*User, error)
	retrieveByEmail(email string, user *User) (*User, error)
	retrieveVerifiedUserByEmail(email string, user *User) (*User, error)
	retrievePage(filters *UserFilters) (*UserPage, error)
	update(id uint, user *User) (*User, error)
}

type DBUserRepository struct {
	dbConfig config.DatabaseConfig
}

func (dbUserRepository *DBUserRepository) retrieveById(id uint, user *User) (*User, error) {
	connDB := dbUserRepository.dbConfig.Begin()
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

func (dbUserRepository *DBUserRepository) retrieveByEmail(email string, user *User) (*User, error) {
	connDB := dbUserRepository.dbConfig.Begin()
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

func (dbUserRepository *DBUserRepository) retrieveVerifiedUserByEmail(email string, user *User) (*User, error) {
	connDB := dbUserRepository.dbConfig.Begin()
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

func (dbUserRepository *DBUserRepository) retrievePage(filters *UserFilters) (*UserPage, error) {
	users := []User{}
	err := dbUserRepository.dbConfig.Begin().
		Model(&model.User{}).
		Scopes(dbUserRepository.dbConfig.Paginate(&filters.Filters)).
		Find(&users)
	if err.Error != nil {
		return nil, err.Error
	}
	return &UserPage{
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

func (dbUserRepository *DBUserRepository) update(id uint, user *User) (*User, error) {
	connDB := dbUserRepository.dbConfig.Begin()
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

func NewDBUser() *DBUserRepository {
	dbConfig := app.Http.Database
	return &DBUserRepository{
		dbConfig: dbConfig,
	}
}
