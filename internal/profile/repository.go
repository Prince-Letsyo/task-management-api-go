// Package profile for user profile repository implementations.
package profile

import (
	"errors"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
	"github.com/Prince-Letsyo/task-management-api-go/internal/model"
	"github.com/Prince-Letsyo/task-management-api-go/internal/types"
	"gorm.io/gorm"
)

type IProfileRepository interface {
	retrieve(userID uint, profile *types.Profile) (*types.Profile, error)
	update(id uint, user *types.Profile) (*types.Profile, error)
}

type DBProfileRepository struct {
	*config.AppCfg
}

func (dbProfileRepository *DBProfileRepository) typeModel(profile *types.Profile) *model.Profile {
	var user *model.User

	if profile.User != nil {
		user = &model.User{
			FirstName: profile.User.FirstName,
			LastName:  profile.User.LastName,
			Email:     profile.User.Email,
			UserName:  profile.User.UserName,
		}
	}
	return &model.Profile{
		UserID:    profile.UserID,
		AvatarURL: profile.AvatarURL,
		Bio:       profile.Bio,
		User:      user,
	}
}

func (dbProfileRepository *DBProfileRepository) modelType(profileModel *model.Profile) *types.Profile {
	var user *types.User

	if profileModel.User != nil {
		user = &types.User{
			ID:        profileModel.User.ID,
			FirstName: profileModel.User.FirstName,
			LastName:  profileModel.User.LastName,
			Email:     profileModel.User.Email,
			UserName:  profileModel.User.UserName,
		}
	}
	return &types.Profile{
		ID:        profileModel.ID,
		UserID:    profileModel.UserID,
		AvatarURL: profileModel.AvatarURL,
		Bio:       profileModel.Bio,
		User:      user,
	}
}

func (dbProfileRepository *DBProfileRepository) retrieve(userID uint, profile *types.Profile) (*types.Profile, error) {
	var profileModel *model.Profile
	if err := dbProfileRepository.Database.Model(&model.Profile{}).Preload("User").
		Where("user_id = ? ", userID).First(profileModel); err.Error != nil {
		if errors.Is(err.Error, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrProfileNotFound
		}
		return nil, err.Error
	}
	convertedProfile := dbProfileRepository.modelType(profileModel)
	if profile != nil {
		*profile = *convertedProfile
	}
	return profile, nil
}

func (dbProfileRepository *DBProfileRepository) update(id uint, profile *types.Profile) (*types.Profile, error) {
	profileModel := dbProfileRepository.typeModel(profile)

	if err := dbProfileRepository.Database.Model(&model.Profile{}).
		Where("id = ? ", id).
		Updates(profileModel).First(profileModel); err.Error != nil {
		if errors.Is(err.Error, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrProfileNotFound
		}
		return nil, err.Error
	}
	profile = dbProfileRepository.modelType(profileModel)
	return profile, nil
}

func NewDBProfileRepository(appCfg *config.AppCfg) IProfileRepository {
	return &DBProfileRepository{AppCfg: appCfg}
}
