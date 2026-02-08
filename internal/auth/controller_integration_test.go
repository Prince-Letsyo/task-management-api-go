package auth

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/model"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) SignUp(userCreate *UserCreateRequest) (*model.User, string, time.Time, error) {
	args := m.Called(userCreate)
	if args.Get(0) == nil {
		return nil, "", time.Time{}, args.Error(3)
	}
	return args.Get(0).(*model.User), args.String(1), args.Get(2).(time.Time), args.Error(3)
}

func (m *MockAuthService) SendActivationEmail(email string) (*model.User, string, time.Time, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, "", time.Time{}, args.Error(3)
	}
	return args.Get(0).(*model.User), args.String(1), args.Get(2).(time.Time), args.Error(3)
}

func (m *MockAuthService) ActivateAccount(token string) (*model.User, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthService) Login(username string, password string, userAgent *string, ipAddress *string) (*UserResponse, error) {
	args := m.Called(username, password, userAgent, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserResponse), args.Error(1)
}

func (m *MockAuthService) Login2FA(tempToken string, totpToken string, userAgent *string, ipAddress *string) (*UserResponse, error) {
	args := m.Called(tempToken, totpToken, userAgent, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserResponse), args.Error(1)
}

func (m *MockAuthService) GetAccessToken(refreshToken string) (*TokenModel, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TokenModel), args.Error(1)
}

func (m *MockAuthService) RequestPasswordReset(email string) (*model.User, string, time.Time, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, "", time.Time{}, args.Error(3)
	}
	return args.Get(0).(*model.User), args.String(1), args.Get(2).(time.Time), args.Error(3)
}

func (m *MockAuthService) ResetPassword(token string, email string, password string) (*model.User, error) {
	args := m.Called(token, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthService) Logout(refreshToken string) error {
	args := m.Called(refreshToken)
	return args.Error(0)
}

func (m *MockAuthService) LogoutAll(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockAuthService) ListSessions(userID uint) ([]model.Session, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Session), args.Error(1)
}

func (m *MockAuthService) ChangeEmail(username string, newEmail string, password string) error {
	args := m.Called(username, newEmail, password)
	return args.Error(0)
}

func (m *MockAuthService) AdminRevokeSessions(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockAuthService) Enable2FA(username string) (string, string, error) {
	args := m.Called(username)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthService) Disable2FA(username string) error {
	args := m.Called(username)
	return args.Error(0)
}

func (m *MockAuthService) ValidateAccessToken(token string) (*AuthClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthClaims), args.Error(1)
}

func TestAuthController_SignIn(t *testing.T) {
	app := fiber.New()
	cfg := &config.AppCfg{
		JwtSecrets: config.JwtSecrets{Secret: "test-secret"},
	}

	mockSvc := new(MockAuthService)
	authC := Auth{router: app}
	newAuthControllerWithService(authC, cfg, mockSvc)

	mockSvc.On("Login", "alice", "Password123!", mock.Anything, mock.Anything).
		Return(&UserResponse{
			Requires2FA: false,
			Token: &TokenModel{
				AccessToken:  AccessToken{Token: "access", Duration: time.Now().Add(time.Minute)},
				RefreshToken: RefreshToken{Token: "refresh", Duration: time.Now().Add(time.Hour)},
			},
		}, nil)

	body := []byte(`{"username":"alice","password":"Password123!"}`)
	req := httptest.NewRequest("POST", "/sign-in", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var payload map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	assert.Contains(t, payload, "requires_2fa")
	assert.Contains(t, payload, "token")

	mockSvc.AssertExpectations(t)
}

func TestAuthController_RefreshRotation(t *testing.T) {
	app := fiber.New()
	cfg := &config.AppCfg{
		JwtSecrets: config.JwtSecrets{Secret: "test-secret"},
	}

	mockSvc := new(MockAuthService)
	authC := Auth{router: app}
	newAuthControllerWithService(authC, cfg, mockSvc)

	mockSvc.On("GetAccessToken", "refresh-token").
		Return(&TokenModel{
			AccessToken:  AccessToken{Token: "access", Duration: time.Now().Add(time.Minute)},
			RefreshToken: RefreshToken{Token: "refresh-2", Duration: time.Now().Add(time.Hour)},
		}, nil)

	body := []byte(`{"token":"refresh-token"}`)
	req := httptest.NewRequest("POST", "/access", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var payload map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	assert.Contains(t, payload, "access_token")
	assert.Contains(t, payload, "refresh_token")

	mockSvc.AssertExpectations(t)
}

func TestAuthController_ListSessions_ValidatesSession(t *testing.T) {
	app := fiber.New()
	cfg := &config.AppCfg{
		JwtSecrets: config.JwtSecrets{Secret: "test-secret"},
	}

	mockSvc := new(MockAuthService)
	authC := Auth{router: app}
	newAuthControllerWithService(authC, cfg, mockSvc)

	mockSvc.On("ValidateAccessToken", "valid-access").
		Return(&AuthClaims{UserID: 1, Username: "alice"}, nil)
	mockSvc.On("ListSessions", uint(1)).
		Return([]model.Session{{ID: "s1"}}, nil)

	req := httptest.NewRequest("GET", "/sessions", nil)
	req.Header.Set("Authorization", "Bearer valid-access")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	mockSvc.AssertExpectations(t)
}

func TestAuthController_ListSessions_Revoked(t *testing.T) {
	app := fiber.New()
	cfg := &config.AppCfg{
		JwtSecrets: config.JwtSecrets{Secret: "test-secret"},
	}

	mockSvc := new(MockAuthService)
	authC := Auth{router: app}
	newAuthControllerWithService(authC, cfg, mockSvc)

	mockSvc.On("ValidateAccessToken", "revoked-access").
		Return(nil, assert.AnError)

	req := httptest.NewRequest("GET", "/sessions", nil)
	req.Header.Set("Authorization", "Bearer revoked-access")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	mockSvc.AssertExpectations(t)
}

func TestAuthController_LogoutAll_ValidatesSession(t *testing.T) {
	app := fiber.New()
	cfg := &config.AppCfg{
		JwtSecrets: config.JwtSecrets{Secret: "test-secret"},
	}

	mockSvc := new(MockAuthService)
	authC := Auth{router: app}
	newAuthControllerWithService(authC, cfg, mockSvc)

	mockSvc.On("ValidateAccessToken", "valid-access").
		Return(&AuthClaims{UserID: 7, Username: "alice"}, nil)
	mockSvc.On("LogoutAll", uint(7)).Return(nil)

	req := httptest.NewRequest("POST", "/logout-all", nil)
	req.Header.Set("Authorization", "Bearer valid-access")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	mockSvc.AssertExpectations(t)
}

func TestAuthController_ChangeEmail_ValidatesSession(t *testing.T) {
	app := fiber.New()
	cfg := &config.AppCfg{
		JwtSecrets: config.JwtSecrets{Secret: "test-secret"},
	}

	mockSvc := new(MockAuthService)
	authC := Auth{router: app}
	newAuthControllerWithService(authC, cfg, mockSvc)

	mockSvc.On("ValidateAccessToken", "valid-access").
		Return(&AuthClaims{UserID: 1, Username: "alice"}, nil)
	mockSvc.On("ChangeEmail", "alice", "alice2@example.com", "Password123!").Return(nil)

	body := []byte(`{"new_email":"alice2@example.com","password":"Password123!"}`)
	req := httptest.NewRequest("POST", "/change-email", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-access")
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	mockSvc.AssertExpectations(t)
}

func TestAuthController_EnableDisable2FA_ValidatesSession(t *testing.T) {
	app := fiber.New()
	cfg := &config.AppCfg{
		JwtSecrets: config.JwtSecrets{Secret: "test-secret"},
	}

	mockSvc := new(MockAuthService)
	authC := Auth{router: app}
	newAuthControllerWithService(authC, cfg, mockSvc)

	mockSvc.On("ValidateAccessToken", "valid-access").
		Return(&AuthClaims{UserID: 1, Username: "alice"}, nil)
	mockSvc.On("Enable2FA", "alice").
		Return("SECRET", "data:image/png;base64,ABC", nil)
	mockSvc.On("Disable2FA", "alice").Return(nil)

	reqEnable := httptest.NewRequest("POST", "/enable-2fa", nil)
	reqEnable.Header.Set("Authorization", "Bearer valid-access")
	respEnable, err := app.Test(reqEnable)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, respEnable.StatusCode)

	reqDisable := httptest.NewRequest("POST", "/disable-2fa", nil)
	reqDisable.Header.Set("Authorization", "Bearer valid-access")
	respDisable, err := app.Test(reqDisable)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, respDisable.StatusCode)

	mockSvc.AssertExpectations(t)
}

func TestAuthController_AdminRevokeSessions_APIKey(t *testing.T) {
	app := fiber.New()
	cfg := &config.AppCfg{
		Auth: config.AuthSettings{AdminAPIKey: "secret-key"},
	}

	mockSvc := new(MockAuthService)
	authC := Auth{router: app}
	newAuthControllerWithService(authC, cfg, mockSvc)

	mockSvc.On("AdminRevokeSessions", uint(9)).Return(nil)

	req := httptest.NewRequest("POST", "/admin/revoke-sessions/9", nil)
	req.Header.Set("X-API-Key", "secret-key")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	mockSvc.AssertExpectations(t)
}

func TestAuthController_AdminRevokeSessions_InvalidAPIKey(t *testing.T) {
	app := fiber.New()
	cfg := &config.AppCfg{
		Auth: config.AuthSettings{AdminAPIKey: "secret-key"},
	}

	mockSvc := new(MockAuthService)
	authC := Auth{router: app}
	newAuthControllerWithService(authC, cfg, mockSvc)

	req := httptest.NewRequest("POST", "/admin/revoke-sessions/9", nil)
	req.Header.Set("X-API-Key", "wrong-key")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
