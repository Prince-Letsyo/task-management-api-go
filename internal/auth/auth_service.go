package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/oarkflow/log"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/model"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
)

type AuthService struct {
	repo   AuthRepository
	auth   config.AuthSettings
	jwt    config.JwtSecrets
	server config.ServerConfig
}

func NewAuthService(repo AuthRepository, auth config.AuthSettings, jwt config.JwtSecrets, server config.ServerConfig) *AuthService {
	return &AuthService{
		repo:   repo,
		auth:   auth,
		jwt:    jwt,
		server: server,
	}
}

type IAuthService interface {
	Login(username string, password string, userAgent *string, ipAddress *string) (*UserResponse, error)
	Login2FA(tempToken string, totpToken string, userAgent *string, ipAddress *string) (*UserResponse, error)
	GetAccessToken(refreshToken string) (*TokenModel, error)
	Logout(refreshToken string) error
	LogoutAll(userID uint) error
	ListSessions(userID uint) ([]model.Session, error)
	AdminRevokeSessions(userID uint) error
	Enable2FA(username string) (string, string, error)
	Disable2FA(username string) error
	ValidateAccessToken(token string) (*AuthClaims, error)
}

func (s *AuthService) Login(username string, password string, userAgent *string, ipAddress *string) (*UserResponse, error) {
	user, err := s.repo.GetUserByUsernameAnyStatus(username)
	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("failed to get user for login")
		return nil, err
	}
	if !user.IsVerified {
		return nil, pkg.NewAppError(fiber.StatusForbidden, "user account is not active")
	}
	match, _ := pkg.DefaultPassword.Match(password, user.Password)
	if !match {
		return nil, pkg.NewAppError(fiber.StatusUnauthorized, "invalid username or password")
	}
	if user.Is2FAEnabled {
		claims := buildClaims(AuthClaims{
			Username:   user.UserName,
			Email:      user.Email,
			UserID:     user.ID,
			TokenType:  "temp_2fa",
			MFAPending: true,
		}, s.auth.Temp2FAExpire)
		token, err := createToken(s.jwt.Secret, claims)
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
	claims, err := parseToken(s.jwt.Secret, tempToken)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "temp_2fa" || !claims.MFAPending {
		return nil, pkg.NewAppError(fiber.StatusUnauthorized, "invalid token type")
	}
	if claims.Username == "" {
		return nil, pkg.NewAppError(fiber.StatusUnauthorized, "invalid token payload")
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
	claims, err := parseToken(s.jwt.Secret, refreshToken)
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
	refreshHash := hashToken(s.jwt.Secret, refreshToken)
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
	}, s.jwt.AccessExpire)
	accessToken, err := createToken(s.jwt.Secret, accessClaims)
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
	}, s.jwt.RefreshExpire)
	newRefreshToken, err := createToken(s.jwt.Secret, refreshClaims)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateSessionToken(session.ID, hashToken(s.jwt.Secret, newRefreshToken)); err != nil {
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

func (s *AuthService) Logout(refreshToken string) error {
	claims, err := parseToken(s.jwt.Secret, refreshToken)
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
	refreshHash := hashToken(s.jwt.Secret, refreshToken)
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
		Issuer:      s.server.Name,
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
	claims, err := parseToken(s.jwt.Secret, token)
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
	}, s.jwt.AccessExpire)
	accessToken, err := createToken(s.jwt.Secret, accessClaims)
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
	}, s.jwt.RefreshExpire)
	refreshToken, err := createToken(s.jwt.Secret, refreshClaims)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateSession(&model.Session{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: hashToken(s.jwt.Secret, refreshToken),
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
	key := s.auth.TOTPSecretKey
	if strings.TrimSpace(key) == "" {
		key = s.jwt.Secret
	}
	return pkg.EncryptStringHexKey(key, secret)
}

func (s *AuthService) decryptTOTPSecret(encrypted string) (string, error) {
	key := s.auth.TOTPSecretKey
	if strings.TrimSpace(key) == "" {
		key = s.jwt.Secret
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
