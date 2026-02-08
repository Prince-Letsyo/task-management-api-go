package user

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/Prince-Letsyo/task-management-api-go/internal/model"
	"github.com/Prince-Letsyo/task-management-api-go/internal/types"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
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
	db *gorm.DB
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
		FirstName:           user.FirstName,
		LastName:            user.LastName,
		Email:               user.Email,
		Password:            user.Password,
		IsVerified:          user.IsVerified,
		IsAdmin:             user.IsAdmin,
		UserName:            user.UserName,
		TOTPSecret:          user.TOTPSecret,
		Is2FAEnabled:        user.Is2FAEnabled,
		RefreshTokenVersion: user.RefreshTokenVersion,
		Profile:             profile,
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
		ID:                  userModel.ID,
		FirstName:           userModel.FirstName,
		LastName:            userModel.LastName,
		Password:            userModel.Password,
		Email:               userModel.Email,
		IsVerified:          userModel.IsVerified,
		IsAdmin:             userModel.IsAdmin,
		UserName:            userModel.UserName,
		TOTPSecret:          userModel.TOTPSecret,
		Is2FAEnabled:        userModel.Is2FAEnabled,
		RefreshTokenVersion: userModel.RefreshTokenVersion,
		Profile:             profile,
	}
}

func (dbUserRepository *DBUserRepository) store(user *types.User) (*types.User, error) {
	userModel := dbUserRepository.typeModel(user)
	if err := dbUserRepository.db.Create(userModel); err.Error != nil {
		return nil, err.Error
	}
	return dbUserRepository.modelType(userModel), nil
}

func (dbUserRepository *DBUserRepository) retrieveByID(id uint, user *types.User) (*types.User, error) {
	var userModel = &model.User{}
	if err := dbUserRepository.db.
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
	if err := dbUserRepository.db.Where(&model.User{Email: email}).
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
	if err := dbUserRepository.db.
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

	// Need a way to paginate without AppCfg. Actually DatabaseConfig has Paginate method
	// For now let's just use the GORM DB directly if Paginate is defined on config.DatabaseConfig
	// Wait, dbUserRepository only has *gorm.DB.

	query := dbUserRepository.db.Preload("Profile").Model(&model.User{})

	// How to use the Paginate scope? It's defined on config.DatabaseConfig.
	// We can't easily access it here without AppCfg.
	// Let's implement a simple version or move Paginate to pkg.

	// For now, let's just do it manually.
	var total int64
	query.Count(&total)
	filters.Total = total

	limit := filters.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := (filters.Offset - 1) * limit
	if offset < 0 {
		offset = 0
	}

	if err := query.Offset(int(offset)).Limit(int(limit)).Find(&userModels); err.Error != nil {
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
	if err := dbUserRepository.db.Model(&model.User{}).
		Preload("Profile").
		Where("id = ? ", id).
		Updates(updatedValues).First(userModel); err.Error != nil {

		if errors.Is(err.Error, gorm.ErrInvalidDB) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, errors.Wrap(err.Error, "cannot update user")
	}
	user = dbUserRepository.modelType(userModel)

	return user, nil
}

func NewDBUser(db *gorm.DB) *DBUserRepository {
	return &DBUserRepository{
		db: db,
	}
}
