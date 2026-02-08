package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/model"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
)

type AuthService struct {
	repo AuthRepository
	cfg  *config.AppCfg
}

func NewAuthService(cfg *config.AppCfg, repo AuthRepository) *AuthService {
	return &AuthService{cfg: cfg, repo: repo}
}

type IAuthService interface {
	SignUp(userCreate *UserCreateRequest) (*model.User, string, time.Time, error)
	SendActivationEmail(email string) (*model.User, string, time.Time, error)
	ActivateAccount(token string) (*model.User, error)
	Login(username string, password string, userAgent *string, ipAddress *string) (*UserResponse, error)
	Login2FA(tempToken string, totpToken string, userAgent *string, ipAddress *string) (*UserResponse, error)
	GetAccessToken(refreshToken string) (*TokenModel, error)
	RequestPasswordReset(email string) (*model.User, string, time.Time, error)
	ResetPassword(token string, email string, password string) (*model.User, error)
	Logout(refreshToken string) error
	LogoutAll(userID uint) error
	ListSessions(userID uint) ([]model.Session, error)
	ChangeEmail(username string, newEmail string, password string) error
	AdminRevokeSessions(userID uint) error
	Enable2FA(username string) (string, string, error)
	Disable2FA(username string) error
	ValidateAccessToken(token string) (*AuthClaims, error)
}

func (s *AuthService) SignUp(userCreate *UserCreateRequest) (*model.User, string, time.Time, error) {
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
		return nil, "", time.Time{}, err
	}
	claims := buildClaims(AuthClaims{
		Username:  created.UserName,
		Email:     created.Email,
		UserID:    created.ID,
		TokenType: "activate",
	}, s.cfg.Auth.ActivateExpire)
	token, err := createToken(s.cfg.JwtSecrets.Secret, claims)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	return created, token, claims.ExpiresAt.Time, nil
}

func (s *AuthService) SendActivationEmail(email string) (*model.User, string, time.Time, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	if user.IsVerified {
		return nil, "", time.Time{}, errors.New("user account is already active")
	}
	claims := buildClaims(AuthClaims{
		Username:  user.UserName,
		Email:     user.Email,
		UserID:    user.ID,
		TokenType: "activate",
	}, s.cfg.Auth.ActivateExpire)
	token, err := createToken(s.cfg.JwtSecrets.Secret, claims)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	return user, token, claims.ExpiresAt.Time, nil
}

func (s *AuthService) ActivateAccount(token string) (*model.User, error) {
	claims, err := parseToken(s.cfg.JwtSecrets.Secret, token)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "activate" {
		return nil, errors.New("invalid token type")
	}
	if claims.Username == "" {
		return nil, errors.New("invalid token payload")
	}
	return s.repo.ActivateUserAccount(claims.Username)
}

func (s *AuthService) Login(username string, password string, userAgent *string, ipAddress *string) (*UserResponse, error) {
	user, err := s.repo.GetUserByUsernameAnyStatus(username)
	if err != nil {
		return nil, err
	}
	if !user.IsVerified {
		return nil, errors.New("user account is not active")
	}
	match, _ := pkg.DefaultPassword.Match(password, user.Password)
	if !match {
		return nil, errors.New("invalid username or password")
	}
	if user.Is2FAEnabled {
		claims := buildClaims(AuthClaims{
			Username:   user.UserName,
			Email:      user.Email,
			UserID:     user.ID,
			TokenType:  "temp_2fa",
			MFAPending: true,
		}, s.cfg.Auth.Temp2FAExpire)
		token, err := createToken(s.cfg.JwtSecrets.Secret, claims)
		if err != nil {
			return nil, err
		}
		return &UserResponse{
			Requires2FA: true,
			Token: Temp2FAToken{
				Token:    token,
				Duration: claims.ExpiresAt.Time,
			},
		}, nil
	}
	token, err := s.issueTokens(user, userAgent, ipAddress)
	if err != nil {
		return nil, err
	}
	return &UserResponse{Requires2FA: false, Token: token}, nil
}

func (s *AuthService) Login2FA(tempToken string, totpToken string, userAgent *string, ipAddress *string) (*UserResponse, error) {
	claims, err := parseToken(s.cfg.JwtSecrets.Secret, tempToken)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "temp_2fa" || !claims.MFAPending {
		return nil, errors.New("invalid token type")
	}
	if claims.Username == "" {
		return nil, errors.New("invalid token payload")
	}
	user, err := s.repo.GetUserByUsernameAnyStatus(claims.Username)
	if err != nil {
		return nil, err
	}
	if !user.Is2FAEnabled {
		return nil, errors.New("2fa is not enabled for this user")
	}
	if user.TOTPSecret == nil {
		return nil, errors.New("2fa secret is missing. please re-enable 2fa")
	}
	secret, err := s.decryptTOTPSecret(*user.TOTPSecret)
	if err != nil {
		// fallback to legacy plaintext
		secret = *user.TOTPSecret
	}
	valid, err := totp.ValidateCustom(totpToken, secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil || !valid {
		return nil, errors.New("invalid totp token")
	}
	if secret != *user.TOTPSecret {
		encrypted, err := s.encryptTOTPSecret(secret)
		if err == nil {
			_, _ = s.repo.Enable2FA(user.UserName, encrypted)
		}
	}
	token, err := s.issueTokens(user, userAgent, ipAddress)
	if err != nil {
		return nil, err
	}
	return &UserResponse{Requires2FA: false, Token: token}, nil
}

func (s *AuthService) GetAccessToken(refreshToken string) (*TokenModel, error) {
	claims, err := parseToken(s.cfg.JwtSecrets.Secret, refreshToken)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "refresh" {
		return nil, errors.New("invalid token type")
	}
	if claims.Username == "" || claims.Email == "" || claims.UserID == 0 || claims.SID == "" {
		return nil, errors.New("invalid token payload")
	}
	user, err := s.repo.GetUserByUsernameAnyStatus(claims.Username)
	if err != nil {
		return nil, err
	}
	if !user.IsVerified || user.ID != claims.UserID || user.Email != claims.Email {
		return nil, errors.New("invalid token")
	}
	if claims.RefreshVersion != user.RefreshTokenVersion {
		return nil, errors.New("invalid token")
	}
	session, err := s.repo.GetSession(claims.SID)
	if err != nil {
		return nil, err
	}
	if session.RevokedAt != nil || session.UserID != user.ID {
		return nil, errors.New("invalid token")
	}
	refreshHash := hashToken(s.cfg.JwtSecrets.Secret, refreshToken)
	if !compareHash(refreshHash, session.RefreshTokenHash) {
		return nil, errors.New("invalid token")
	}

	accessClaims := buildClaims(AuthClaims{
		Username:       user.UserName,
		Email:          user.Email,
		UserID:         user.ID,
		SID:            session.ID,
		RefreshVersion: user.RefreshTokenVersion,
		TokenType:      "access",
	}, s.cfg.JwtSecrets.AccessExpire)
	accessToken, err := createToken(s.cfg.JwtSecrets.Secret, accessClaims)
	if err != nil {
		return nil, err
	}
	refreshClaims := buildClaims(AuthClaims{
		Username:       user.UserName,
		Email:          user.Email,
		UserID:         user.ID,
		SID:            session.ID,
		RefreshVersion: user.RefreshTokenVersion,
		TokenType:      "refresh",
		JTI:            uuid.NewString(),
	}, s.cfg.JwtSecrets.RefreshExpire)
	newRefreshToken, err := createToken(s.cfg.JwtSecrets.Secret, refreshClaims)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateSessionToken(session.ID, hashToken(s.cfg.JwtSecrets.Secret, newRefreshToken)); err != nil {
		return nil, err
	}
	return &TokenModel{
		AccessToken: AccessToken{
			Token:    accessToken,
			Duration: accessClaims.ExpiresAt.Time,
		},
		RefreshToken: RefreshToken{
			Token:    newRefreshToken,
			Duration: refreshClaims.ExpiresAt.Time,
		},
	}, nil
}

func (s *AuthService) RequestPasswordReset(email string) (*model.User, string, time.Time, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	claims := buildClaims(AuthClaims{
		Username:  user.UserName,
		Email:     user.Email,
		UserID:    user.ID,
		TokenType: "password_reset",
	}, s.cfg.Auth.PasswordResetExpire)
	token, err := createToken(s.cfg.JwtSecrets.Secret, claims)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	return user, token, claims.ExpiresAt.Time, nil
}

func (s *AuthService) ResetPassword(token string, email string, password string) (*model.User, error) {
	claims, err := parseToken(s.cfg.JwtSecrets.Secret, token)
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

func (s *AuthService) Logout(refreshToken string) error {
	claims, err := parseToken(s.cfg.JwtSecrets.Secret, refreshToken)
	if err != nil {
		return err
	}
	if claims.TokenType != "refresh" {
		return errors.New("invalid token type")
	}
	if claims.SID == "" {
		return errors.New("invalid token payload")
	}
	session, err := s.repo.GetSession(claims.SID)
	if err != nil {
		return err
	}
	if session.RevokedAt != nil {
		return nil
	}
	refreshHash := hashToken(s.cfg.JwtSecrets.Secret, refreshToken)
	if !compareHash(refreshHash, session.RefreshTokenHash) {
		return errors.New("invalid token")
	}
	return s.repo.RevokeSession(session.ID)
}

func (s *AuthService) LogoutAll(userID uint) error {
	return s.repo.RevokeUserSessions(userID)
}

func (s *AuthService) ListSessions(userID uint) ([]model.Session, error) {
	return s.repo.ListSessions(userID)
}

func (s *AuthService) ChangeEmail(username string, newEmail string, password string) error {
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

func (s *AuthService) AdminRevokeSessions(userID uint) error {
	return s.repo.RevokeUserSessions(userID)
}

func (s *AuthService) Enable2FA(username string) (string, string, error) {
	user, err := s.repo.GetUserByUsernameAnyStatus(username)
	if err != nil {
		return "", "", err
	}
	if user.Is2FAEnabled {
		return "", "", errors.New("2fa is already enabled for this user")
	}
	secret, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.cfg.Server.Name,
		AccountName: user.Email,
	})
	if err != nil {
		return "", "", err
	}
	plainSecret := secret.Secret()
	encrypted, err := s.encryptTOTPSecret(plainSecret)
	if err != nil {
		return "", "", err
	}
	if _, err := s.repo.Enable2FA(user.UserName, encrypted); err != nil {
		return "", "", err
	}
	if err := s.repo.IncrementRefreshVersion(user.ID); err != nil {
		return "", "", err
	}
	qrCode, err := s.qrCodeDataURI(secret.URL())
	if err != nil {
		return "", "", err
	}
	_ = s.repo.RevokeUserSessions(user.ID)
	return plainSecret, qrCode, nil
}

func (s *AuthService) Disable2FA(username string) error {
	user, err := s.repo.GetUserByUsernameAnyStatus(username)
	if err != nil {
		return err
	}
	if !user.Is2FAEnabled {
		return errors.New("2fa is not enabled for this user")
	}
	if _, err := s.repo.Disable2FA(username); err != nil {
		return err
	}
	if err := s.repo.IncrementRefreshVersion(user.ID); err != nil {
		return err
	}
	return s.repo.RevokeUserSessions(user.ID)
}

func (s *AuthService) ValidateAccessToken(token string) (*AuthClaims, error) {
	claims, err := parseToken(s.cfg.JwtSecrets.Secret, token)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "access" {
		return nil, errors.New("invalid token type")
	}
	if claims.SID == "" || claims.UserID == 0 || claims.Username == "" {
		return nil, errors.New("invalid token payload")
	}
	session, err := s.repo.GetSession(claims.SID)
	if err != nil {
		return nil, err
	}
	if session.RevokedAt != nil || session.UserID != claims.UserID {
		return nil, errors.New("invalid token")
	}
	user, err := s.repo.GetUserByUsernameAnyStatus(claims.Username)
	if err != nil {
		return nil, err
	}
	if user.ID != claims.UserID || user.Email != claims.Email {
		return nil, errors.New("invalid token")
	}
	if claims.RefreshVersion != user.RefreshTokenVersion {
		return nil, errors.New("invalid token")
	}
	_ = s.repo.UpdateSessionLastUsed(session.ID)
	return claims, nil
}

func (s *AuthService) issueTokens(user *model.User, userAgent *string, ipAddress *string) (*TokenModel, error) {
	sessionID := uuid.NewString()
	accessClaims := buildClaims(AuthClaims{
		Username:       user.UserName,
		Email:          user.Email,
		UserID:         user.ID,
		SID:            sessionID,
		RefreshVersion: user.RefreshTokenVersion,
		TokenType:      "access",
	}, s.cfg.JwtSecrets.AccessExpire)
	accessToken, err := createToken(s.cfg.JwtSecrets.Secret, accessClaims)
	if err != nil {
		return nil, err
	}
	refreshClaims := buildClaims(AuthClaims{
		Username:       user.UserName,
		Email:          user.Email,
		UserID:         user.ID,
		SID:            sessionID,
		RefreshVersion: user.RefreshTokenVersion,
		TokenType:      "refresh",
		JTI:            uuid.NewString(),
	}, s.cfg.JwtSecrets.RefreshExpire)
	refreshToken, err := createToken(s.cfg.JwtSecrets.Secret, refreshClaims)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateSession(&model.Session{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: hashToken(s.cfg.JwtSecrets.Secret, refreshToken),
		UserAgent:        userAgent,
		IPAddress:        ipAddress,
		LastUsedAt:       nil,
	}); err != nil {
		return nil, err
	}
	return &TokenModel{
		AccessToken: AccessToken{
			Token:    accessToken,
			Duration: accessClaims.ExpiresAt.Time,
		},
		RefreshToken: RefreshToken{
			Token:    refreshToken,
			Duration: refreshClaims.ExpiresAt.Time,
		},
	}, nil
}

func (s *AuthService) encryptTOTPSecret(secret string) (string, error) {
	key := s.cfg.Auth.TOTPSecretKey
	if strings.TrimSpace(key) == "" {
		key = s.cfg.JwtSecrets.Secret
	}
	return pkg.EncryptStringHexKey(key, secret)
}

func (s *AuthService) decryptTOTPSecret(encrypted string) (string, error) {
	key := s.cfg.Auth.TOTPSecretKey
	if strings.TrimSpace(key) == "" {
		key = s.cfg.JwtSecrets.Secret
	}
	return pkg.DecryptStringHexKey(key, encrypted)
}

func (s *AuthService) qrCodeDataURI(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" {
		return "", fmt.Errorf("invalid otp url")
	}
	png, err := GenerateQRCodePNG(rawURL)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(png)
	return "data:image/png;base64," + encoded, nil
}
