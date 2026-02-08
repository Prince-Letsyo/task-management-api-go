package auth

import (
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/model"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
)

type AuthRepository interface {
	CreateUser(user *model.User) (*model.User, error)
	GetUserByEmail(email string) (*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
	GetUserByUsernameAnyStatus(username string) (*model.User, error)
	ActivateUserAccount(username string) (*model.User, error)
	UpdateUserPasswordByEmail(email string, newPassword string) (*model.User, error)
	UpdateUserEmail(userID uint, newEmail string) (*model.User, error)
	Enable2FA(username string, totpSecret string) (*model.User, error)
	Disable2FA(username string) (*model.User, error)
	IncrementRefreshVersion(userID uint) error

	CreateSession(session *model.Session) error
	GetSession(sessionID string) (*model.Session, error)
	UpdateSessionToken(sessionID string, refreshTokenHash string) error
	UpdateSessionLastUsed(sessionID string) error
	RevokeSession(sessionID string) error
	RevokeUserSessions(userID uint) error
	ListSessions(userID uint) ([]model.Session, error)
}

type DBAuthRepository struct {
	*config.AppCfg
}

func NewDBAuthRepository(appCfg *config.AppCfg) *DBAuthRepository {
	return &DBAuthRepository{AppCfg: appCfg}
}

func (r *DBAuthRepository) CreateUser(user *model.User) (*model.User, error) {
	if err := r.Database.Create(user); err.Error != nil {
		return nil, err.Error
	}
	return user, nil
}

func (r *DBAuthRepository) GetUserByEmail(email string) (*model.User, error) {
	user := &model.User{}
	if err := r.Database.Where(&model.User{Email: email}).First(user); err.Error != nil {
		if errors.Is(err.Error, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, errors.Wrap(err.Error, "failed to retrieve user by email")
	}
	return user, nil
}

func (r *DBAuthRepository) GetUserByUsername(username string) (*model.User, error) {
	user := &model.User{}
	if err := r.Database.Where(&model.User{UserName: username, IsVerified: true}).First(user); err.Error != nil {
		if errors.Is(err.Error, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, errors.Wrap(err.Error, "failed to retrieve active user by username")
	}
	return user, nil
}

func (r *DBAuthRepository) GetUserByUsernameAnyStatus(username string) (*model.User, error) {
	user := &model.User{}
	if err := r.Database.Where(&model.User{UserName: username}).First(user); err.Error != nil {
		if errors.Is(err.Error, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, errors.Wrap(err.Error, "failed to retrieve user by username")
	}
	return user, nil
}

func (r *DBAuthRepository) ActivateUserAccount(username string) (*model.User, error) {
	user := &model.User{}
	if err := r.Database.Where(&model.User{UserName: username}).First(user); err.Error != nil {
		if errors.Is(err.Error, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, err.Error
	}
	user.IsVerified = true
	if err := r.Database.Save(user); err.Error != nil {
		return nil, err.Error
	}
	return user, nil
}

func (r *DBAuthRepository) UpdateUserPasswordByEmail(email string, newPassword string) (*model.User, error) {
	user := &model.User{}
	if err := r.Database.Where(&model.User{Email: email}).First(user); err.Error != nil {
		if errors.Is(err.Error, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, err.Error
	}
	user.Password = newPassword
	if err := r.Database.Save(user); err.Error != nil {
		return nil, err.Error
	}
	return user, nil
}

func (r *DBAuthRepository) UpdateUserEmail(userID uint, newEmail string) (*model.User, error) {
	user := &model.User{}
	if err := r.Database.First(user, userID); err.Error != nil {
		if errors.Is(err.Error, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, err.Error
	}
	user.Email = newEmail
	if err := r.Database.Save(user); err.Error != nil {
		return nil, err.Error
	}
	return user, nil
}

func (r *DBAuthRepository) Enable2FA(username string, totpSecret string) (*model.User, error) {
	user := &model.User{}
	if err := r.Database.Where(&model.User{UserName: username}).First(user); err.Error != nil {
		if errors.Is(err.Error, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, err.Error
	}
	user.Is2FAEnabled = true
	user.TOTPSecret = &totpSecret
	if err := r.Database.Save(user); err.Error != nil {
		return nil, err.Error
	}
	return user, nil
}

func (r *DBAuthRepository) Disable2FA(username string) (*model.User, error) {
	user := &model.User{}
	if err := r.Database.Where(&model.User{UserName: username}).First(user); err.Error != nil {
		if errors.Is(err.Error, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, err.Error
	}
	user.Is2FAEnabled = false
	user.TOTPSecret = nil
	if err := r.Database.Save(user); err.Error != nil {
		return nil, err.Error
	}
	return user, nil
}

func (r *DBAuthRepository) IncrementRefreshVersion(userID uint) error {
	if err := r.Database.Model(&model.User{}).
		Where("id = ?", userID).
		UpdateColumn("refresh_token_version", gorm.Expr("refresh_token_version + 1")); err.Error != nil {
		return err.Error
	}
	return nil
}

func (r *DBAuthRepository) CreateSession(session *model.Session) error {
	if err := r.Database.Create(session); err.Error != nil {
		return err.Error
	}
	return nil
}

func (r *DBAuthRepository) GetSession(sessionID string) (*model.Session, error) {
	session := &model.Session{}
	if err := r.Database.Where(&model.Session{ID: sessionID}).First(session); err.Error != nil {
		if errors.Is(err.Error, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, err.Error
	}
	return session, nil
}

func (r *DBAuthRepository) UpdateSessionToken(sessionID string, refreshTokenHash string) error {
	now := time.Now().UTC()
	if err := r.Database.Model(&model.Session{}).
		Where("id = ?", sessionID).
		Updates(map[string]interface{}{
			"refresh_token_hash": refreshTokenHash,
			"last_used_at":       &now,
		}); err.Error != nil {
		return err.Error
	}
	return nil
}

func (r *DBAuthRepository) UpdateSessionLastUsed(sessionID string) error {
	now := time.Now().UTC()
	if err := r.Database.Model(&model.Session{}).
		Where("id = ?", sessionID).
		Update("last_used_at", &now); err.Error != nil {
		return err.Error
	}
	return nil
}

func (r *DBAuthRepository) RevokeSession(sessionID string) error {
	now := time.Now().UTC()
	if err := r.Database.Model(&model.Session{}).
		Where("id = ?", sessionID).
		Update("revoked_at", &now); err.Error != nil {
		return err.Error
	}
	return nil
}

func (r *DBAuthRepository) RevokeUserSessions(userID uint) error {
	now := time.Now().UTC()
	if err := r.Database.Model(&model.Session{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", &now); err.Error != nil {
		return err.Error
	}
	return nil
}

func (r *DBAuthRepository) ListSessions(userID uint) ([]model.Session, error) {
	sessions := []model.Session{}
	if err := r.Database.Where("user_id = ?", userID).Order("created_at DESC").Find(&sessions); err.Error != nil {
		return nil, err.Error
	}
	return sessions, nil
}
