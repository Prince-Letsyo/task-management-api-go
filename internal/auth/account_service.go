package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/oarkflow/log"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/model"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
)

type IAccountService interface {
	SignUp(userCreate *UserCreateRequest) (*model.User, string, time.Time, error)
	SendActivationEmail(email string) (*model.User, string, time.Time, error)
	ActivateAccount(token string) (*model.User, error)
	RequestPasswordReset(email string) (*model.User, string, time.Time, error)
	ResetPassword(token string, email string, password string) (*model.User, error)
	ChangeEmail(username string, newEmail string, password string) error
}

type AccountService struct {
	repo AuthRepository
	auth config.AuthSettings
	jwt  config.JwtSecrets
}

func NewAccountService(repo AuthRepository, auth config.AuthSettings, jwt config.JwtSecrets) *AccountService {
	return &AccountService{
		repo: repo,
		auth: auth,
		jwt:  jwt,
	}
}

func (s *AccountService) SignUp(userCreate *UserCreateRequest) (*model.User, string, time.Time, error) {
	user := &model.User{
		FirstName:  userCreate.FirstName,
		LastName:   userCreate.LastName,
		UserName:   userCreate.UserName,
		Email:      userCreate.Email,
		Password:   userCreate.Password,
		IsVerified: false,
	}
	created, err := s.repo.CreateUser(user)
	if err != nil {
		log.Error().Err(err).Str("username", userCreate.UserName).Msg("failed to create user during signup")
		return nil, "", time.Time{}, err
	}
	log.Info().Uint("user_id", created.ID).Str("event", "signup_success").Msg("user signed up successfully")
	claims := buildClaims(AuthClaims{
		Username:  created.UserName,
		Email:     created.Email,
		UserID:    created.ID,
		TokenType: "activate",
	}, s.auth.ActivateExpire)
	token, err := createToken(s.jwt.Secret, claims)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	return created, token, claims.ExpiresAt.Time, nil
}

func (s *AccountService) SendActivationEmail(email string) (*model.User, string, time.Time, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	if user.IsVerified {
		return nil, "", time.Time{}, pkg.NewAppError(fiber.StatusBadRequest, "user account is already active")
	}
	claims := buildClaims(AuthClaims{
		Username:  user.UserName,
		Email:     user.Email,
		UserID:    user.ID,
		TokenType: "activate",
	}, s.auth.ActivateExpire)
	token, err := createToken(s.jwt.Secret, claims)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	return user, token, claims.ExpiresAt.Time, nil
}

func (s *AccountService) ActivateAccount(token string) (*model.User, error) {
	claims, err := parseToken(s.jwt.Secret, token)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "activate" {
		return nil, pkg.NewAppError(fiber.StatusUnauthorized, "invalid token type")
	}
	if claims.Username == "" {
		return nil, pkg.NewAppError(fiber.StatusUnauthorized, "invalid token payload")
	}
	return s.repo.ActivateUserAccount(claims.Username)
}

func (s *AccountService) RequestPasswordReset(email string) (*model.User, string, time.Time, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	claims := buildClaims(AuthClaims{
		Username:  user.UserName,
		Email:     user.Email,
		UserID:    user.ID,
		TokenType: "password_reset",
	}, s.auth.PasswordResetExpire)
	token, err := createToken(s.jwt.Secret, claims)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	return user, token, claims.ExpiresAt.Time, nil
}

func (s *AccountService) ResetPassword(token string, email string, password string) (*model.User, error) {
	claims, err := parseToken(s.jwt.Secret, token)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "password_reset" {
		return nil, errors.New("invalid token type")
	}
	if claims.Email == "" || claims.Username == "" {
		return nil, errors.New("invalid token payload")
	}
	if strings.ToLower(claims.Email) != strings.ToLower(email) {
		return nil, errors.New("invalid token payload: email mismatch")
	}
	validation := pkg.DefaultValidator.ValidatePassword(password, claims.Username, claims.Email)
	if !validation.IsValid {
		return nil, errors.New(validation.Errors[0])
	}
	user, err := s.repo.UpdateUserPasswordByEmail(email, password)
	if err != nil {
		return nil, err
	}
	if err := s.repo.IncrementRefreshVersion(user.ID); err != nil {
		return nil, err
	}
	_ = s.repo.RevokeUserSessions(user.ID)
	return user, nil
}

func (s *AccountService) ChangeEmail(username string, newEmail string, password string) error {
	user, err := s.repo.GetUserByUsernameAnyStatus(username)
	if err != nil {
		return err
	}
	match, _ := pkg.DefaultPassword.Match(password, user.Password)
	if !match {
		return errors.New("incorrect username or password")
	}
	if _, err := s.repo.UpdateUserEmail(user.ID, newEmail); err != nil {
		return err
	}
	if err := s.repo.IncrementRefreshVersion(user.ID); err != nil {
		return err
	}
	return s.repo.RevokeUserSessions(user.ID)
}
