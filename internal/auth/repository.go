// Package auth hdsfhhj
package auth

import (
	"errors"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/model"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/types"
	"gorm.io/gorm"
)

type IRegisterRepository interface {
	store(register *types.RegisterForm) (*types.RegisterForm, error)
	updatePassword(id uint, passwordReset *types.PasswordResetForm) (*types.RegisterForm, error)
}

type DBRegisterRepository struct {
	*config.AppCfg
	userService types.IUserService
}

// NewDBRegisterRepository fdg
func NewDBRegisterRepository(userService types.IUserService, appConfig *config.AppCfg) *DBRegisterRepository {
	return &DBRegisterRepository{
		AppCfg:      appConfig,
		userService: userService,
	}
}

func hashPswd(register *types.RegisterForm, hash config.Hash) (err error) {
	if register.CPassword != register.Password {
		return errors.New("passwords do not match")
	}
	hashPas, err := hash.Create(register.Password)
	if err != nil {
		return err
	}
	register.Password = hashPas

	return nil
}

func (dbRegisterRepository *DBRegisterRepository) store(register *types.RegisterForm) (*types.RegisterForm, error) {
	registerModel := &types.RegisterForm{
		FirstName: register.FirstName,
		LastName:  register.LastName,
		Email:     register.Email,
		Password:  register.Password,
		CPassword: register.Password,
	}
	user := &types.User{}

	u, _ := dbRegisterRepository.userService.ViewByEmail(register.Email, user)
	if u != nil {
		return nil, errors.New("user already exists")
	}
	if errHash := hashPswd(register, dbRegisterRepository.Hash); errHash != nil {
		return nil, errHash
	}
	registerModel.Password = register.Password
	connDB := dbRegisterRepository.Database.Begin()
	defer connDB.Commit()
	if err := connDB.Create(registerModel); err.Error != nil {
		connDB.Rollback()
		return nil, err.Error
	}

	return &types.RegisterForm{
		ID:        registerModel.ID,
		FirstName: registerModel.FirstName,
		LastName:  registerModel.LastName,
		Email:     registerModel.Email,
		Password:  registerModel.Password,
		CPassword: registerModel.CPassword,
	}, nil
}

func (dbRegisterRepository *DBRegisterRepository) updatePassword(id uint, passwordReset *types.PasswordResetForm) (*types.RegisterForm, error) {
	connDB := dbRegisterRepository.Database.Begin()
	defer connDB.Commit()
	register := &types.RegisterForm{
		Password:  passwordReset.Password,
		CPassword: passwordReset.CPassword,
	}
	if errHash := hashPswd(register, dbRegisterRepository.Hash); errHash != nil {
		return nil, errHash
	}
	if err := connDB.Model(&model.RegisterForm{}).Where("id = ? ", id).Updates(register); err.Error != nil {
		defer connDB.Rollback()
		if errors.Is(err.Error, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, err.Error
	}
	return register, nil
}
