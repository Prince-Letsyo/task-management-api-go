package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/model"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
)

type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) CreateUser(user *model.User) (*model.User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthRepository) GetUserByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthRepository) GetUserByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthRepository) GetUserByUsernameAnyStatus(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthRepository) ActivateUserAccount(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthRepository) UpdateUserPasswordByEmail(email string, newPassword string) (*model.User, error) {
	args := m.Called(email, newPassword)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthRepository) UpdateUserEmail(userID uint, newEmail string) (*model.User, error) {
	args := m.Called(userID, newEmail)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthRepository) Enable2FA(username string, totpSecret string) (*model.User, error) {
	args := m.Called(username, totpSecret)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthRepository) Disable2FA(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthRepository) IncrementRefreshVersion(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockAuthRepository) CreateSession(session *model.Session) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockAuthRepository) GetSession(sessionID string) (*model.Session, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Session), args.Error(1)
}

func (m *MockAuthRepository) UpdateSessionToken(sessionID string, refreshTokenHash string) error {
	args := m.Called(sessionID, refreshTokenHash)
	return args.Error(0)
}

func (m *MockAuthRepository) UpdateSessionLastUsed(sessionID string) error {
	args := m.Called(sessionID)
	return args.Error(0)
}

func (m *MockAuthRepository) RevokeSession(sessionID string) error {
	args := m.Called(sessionID)
	return args.Error(0)
}

func (m *MockAuthRepository) RevokeUserSessions(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockAuthRepository) ListSessions(userID uint) ([]model.Session, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Session), args.Error(1)
}

func authTestConfig() (config.AuthSettings, config.JwtSecrets, config.ServerConfig) {
	auth := config.AuthSettings{
		Temp2FAExpire:       5 * time.Minute,
		ActivateExpire:      15 * time.Minute,
		PasswordResetExpire: 15 * time.Minute,
	}
	jwt := config.JwtSecrets{
		Secret:        "test-secret-1234567890abcdef",
		AccessExpire:  15 * time.Minute,
		RefreshExpire: 24 * time.Hour,
	}
	server := config.ServerConfig{Name: "TestServer"}
	return auth, jwt, server
}

func TestAuthService_Login_No2FA(t *testing.T) {
	auth, jwt, server := authTestConfig()
	repo := new(MockAuthRepository)
	service := NewAuthService(repo, auth, jwt, server)

	hashed, _ := pkg.DefaultPassword.Create("Password123!")
	user := &model.User{
		Model:        gorm.Model{ID: 1},
		UserName:     "alice",
		Email:        "alice@example.com",
		Password:     hashed,
		IsVerified:   true,
		Is2FAEnabled: false,
	}

	repo.On("GetUserByUsernameAnyStatus", "alice").Return(user, nil)
	repo.On("CreateSession", mock.MatchedBy(func(s *model.Session) bool {
		return s.UserID == user.ID && s.RefreshTokenHash != ""
	})).Return(nil)

	resp, err := service.Login("alice", "Password123!", nil, nil)
	assert.NoError(t, err)
	assert.False(t, resp.Requires2FA)

	tokenModel, ok := resp.Token.(*TokenModel)
	assert.True(t, ok)
	assert.NotEmpty(t, tokenModel.AccessToken.Token)
	assert.NotEmpty(t, tokenModel.RefreshToken.Token)

	accessClaims, err := parseToken(jwt.Secret, tokenModel.AccessToken.Token)
	assert.NoError(t, err)
	refreshClaims, err := parseToken(jwt.Secret, tokenModel.RefreshToken.Token)
	assert.NoError(t, err)
	assert.Equal(t, "access", accessClaims.TokenType)
	assert.Equal(t, "refresh", refreshClaims.TokenType)
	assert.Equal(t, accessClaims.SID, refreshClaims.SID)

	repo.AssertExpectations(t)
}

func TestAuthService_Login_With2FA(t *testing.T) {
	auth, jwt, server := authTestConfig()
	repo := new(MockAuthRepository)
	service := NewAuthService(repo, auth, jwt, server)

	hashed, _ := pkg.DefaultPassword.Create("Password123!")
	user := &model.User{
		Model:        gorm.Model{ID: 1},
		UserName:     "alice",
		Email:        "alice@example.com",
		Password:     hashed,
		IsVerified:   true,
		Is2FAEnabled: true,
	}

	repo.On("GetUserByUsernameAnyStatus", "alice").Return(user, nil)

	resp, err := service.Login("alice", "Password123!", nil, nil)
	assert.NoError(t, err)
	assert.True(t, resp.Requires2FA)

	tempToken, ok := resp.Token.(Temp2FAToken)
	assert.True(t, ok)
	assert.NotEmpty(t, tempToken.Token)

	claims, err := parseToken(jwt.Secret, tempToken.Token)
	assert.NoError(t, err)
	assert.Equal(t, "temp_2fa", claims.TokenType)
	assert.True(t, claims.MFAPending)

	repo.AssertExpectations(t)
}

func TestAuthService_GetAccessToken_Rotation(t *testing.T) {
	auth, jwt, server := authTestConfig()
	repo := new(MockAuthRepository)
	service := NewAuthService(repo, auth, jwt, server)

	user := &model.User{
		Model:      gorm.Model{ID: 1},
		UserName:   "alice",
		Email:      "alice@example.com",
		IsVerified: true,
	}

	sessionID := "session-123"
	refreshClaims := buildClaims(AuthClaims{
		Username:  user.UserName,
		Email:     user.Email,
		UserID:    user.ID,
		SID:       sessionID,
		TokenType: "refresh",
		JTI:       "jti-1",
	}, jwt.RefreshExpire)
	refreshToken, _ := createToken(jwt.Secret, refreshClaims)

	session := &model.Session{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: hashToken(jwt.Secret, refreshToken),
	}

	repo.On("GetUserByUsernameAnyStatus", "alice").Return(user, nil)
	repo.On("GetSession", sessionID).Return(session, nil)
	repo.On("UpdateSessionToken", sessionID, mock.AnythingOfType("string")).Return(nil)

	tokenModel, err := service.GetAccessToken(refreshToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenModel.AccessToken.Token)
	assert.NotEmpty(t, tokenModel.RefreshToken.Token)
	assert.NotEqual(t, refreshToken, tokenModel.RefreshToken.Token)

	accessClaims, err := parseToken(jwt.Secret, tokenModel.AccessToken.Token)
	assert.NoError(t, err)
	assert.Equal(t, "access", accessClaims.TokenType)

	repo.AssertExpectations(t)
}

func TestAuthService_ValidateAccessToken_RefreshVersionMismatch(t *testing.T) {
	auth, jwt, server := authTestConfig()
	repo := new(MockAuthRepository)
	service := NewAuthService(repo, auth, jwt, server)

	user := &model.User{
		Model:               gorm.Model{ID: 1},
		UserName:            "alice",
		Email:               "alice@example.com",
		IsVerified:          true,
		RefreshTokenVersion: 2,
	}

	accessClaims := buildClaims(AuthClaims{
		Username:       user.UserName,
		Email:          user.Email,
		UserID:         user.ID,
		SID:            "session-1",
		RefreshVersion: 1,
		TokenType:      "access",
	}, jwt.AccessExpire)
	accessToken, _ := createToken(jwt.Secret, accessClaims)

	repo.On("GetSession", "session-1").Return(&model.Session{
		ID:     "session-1",
		UserID: user.ID,
	}, nil)
	repo.On("GetUserByUsernameAnyStatus", "alice").Return(user, nil)

	claims, err := service.ValidateAccessToken(accessToken)
	assert.Error(t, err)
	assert.Nil(t, claims)

	repo.AssertExpectations(t)
}
