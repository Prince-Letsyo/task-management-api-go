package auth

import (
	"errors"

	"github.com/Prince-Letsyo/task-management-api-go/cmd/app"
	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/user"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
	"github.com/Prince-Letsyo/task-management-api-go/pkg/model"
	"gorm.io/gorm"
)

type IRegisterRepository interface {
	store(register *RegisterForm) (*RegisterForm, error)
	updatePassword(id uint, passwordReset *PasswordResetForm) (*RegisterForm, error)
}

type DBRegisterRepository struct {
	dbConfig    config.DatabaseConfig
	userService user.IUserService
}

func NewDBRegisterRepository(userService user.IUserService) *DBRegisterRepository {
	dbConfig := app.Http.Database
	return &DBRegisterRepository{
		dbConfig:    dbConfig,
		userService: userService,
	}
}

func hashPswd(register *RegisterForm) (err error) {
	if register.CPassword != register.Password {
		return errors.New("passwords do not match")
	}
	hashPas, err := app.Http.Hash.Create(register.Password)
	if err != nil {
		return err
	}
	register.Password = hashPas

	return nil
}

func (dbRegisterRepository *DBRegisterRepository) store(register *RegisterForm) (*RegisterForm, error) {
	registerModel := &model.RegisterForm{
		FirstName: register.FirstName,
		LastName:  register.LastName,
		Email:     register.Email,
		Password:  register.Password,
		CPassword: register.Password,
	}
	user := &user.User{}

	u, _ := dbRegisterRepository.userService.ViewByEmail(register.Email, user)
	if u != nil {
		return nil, errors.New("user already exists")
	}
	if errHash := hashPswd(register); errHash != nil {
		return nil, errHash
	}
	registerModel.Password = register.Password
	connDB := dbRegisterRepository.dbConfig.Begin()
	defer connDB.Commit()
	if err := connDB.Create(registerModel); err.Error != nil {
		connDB.Rollback()
		return nil, err.Error
	}

	return &RegisterForm{
		ID:        registerModel.ID,
		FirstName: registerModel.FirstName,
		LastName:  registerModel.LastName,
		Email:     registerModel.Email,
		Password:  registerModel.Password,
		CPassword: registerModel.CPassword,
	}, nil
}

func (dbRegisterRepository *DBRegisterRepository) updatePassword(id uint, passwordReset *PasswordResetForm) (*RegisterForm, error) {
	connDB := dbRegisterRepository.dbConfig.Begin()
	defer connDB.Commit()
	register := &RegisterForm{
		Password:  passwordReset.Password,
		CPassword: passwordReset.CPassword,
	}
	if errHash := hashPswd(register); errHash != nil {
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
