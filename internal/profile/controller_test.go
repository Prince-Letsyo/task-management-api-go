package profile

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/auth"
	"github.com/Prince-Letsyo/task-management-api-go/internal/model"
	"github.com/Prince-Letsyo/task-management-api-go/internal/types"
)

// MockProfileService is a mock implementation of IProfileService
type MockProfileService struct {
	mock.Mock
}

func (m *MockProfileService) View(userID uint, profile *types.Profile) (*types.Profile, error) {
	args := m.Called(userID, profile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Profile), args.Error(1)
}

func (m *MockProfileService) Modify(id uint, profile *types.Profile) (*types.Profile, error) {
	args := m.Called(id, profile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Profile), args.Error(1)
}

// MockUserRepo for satisfying controller dependency
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) store(user *types.User) (*types.User, error)                 { return nil, nil }
func (m *MockUserRepo) retrieveByID(id uint, user *types.User) (*types.User, error) { return nil, nil }
func (m *MockUserRepo) retrieveByEmail(email string, user *types.User) (*types.User, error) {
	return nil, nil
}
func (m *MockUserRepo) retrieveVerifiedUserByEmail(email string, user *types.User) (*types.User, error) {
	return nil, nil
}
func (m *MockUserRepo) retrievePage(filters *types.UserFilters) (*types.UserPage, error) {
	return nil, nil
}
func (m *MockUserRepo) update(id uint, updatedValues map[string]interface{}) (*types.User, error) {
	return nil, nil
}

func TestProfileController_GetProfile(t *testing.T) {
	// Setup
	app := fiber.New()
	store := session.New()
	appCfg := &config.AppCfg{
		Session: config.SessionConfig{Store: store},
		JwtSecrets: config.JwtSecrets{
			Secret:       "test-secret",
			AccessExpire: time.Hour,
		},
	}

	mockProfileSvc := new(MockProfileService)
	// Use a dummy user service to satisfy the interface.
	dummyUserSvc := &DummyUserService{}

	userProfile := Profile{
		router:         app,
		profileService: mockProfileSvc,
		userService:    dummyUserSvc,
		authService:    &DummyAuthService{},
	}

	// Init Controller
	newProfileController(userProfile, appCfg)

	// Prepare Token
	userID := uint(1)
	claims := auth.AuthClaims{
		Username:  "testuser",
		Email:     "test@test.com",
		UserID:    userID,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(appCfg.JwtSecrets.Secret))

	// Prepare Request
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{Name: "access-token", Value: token})
	req.Header.Set("Authorization", "Bearer "+token)

	// Expectations
	bio := "My Bio"
	expectedProfile := &types.Profile{UserID: userID, Bio: &bio}
	mockProfileSvc.On("View", userID, mock.AnythingOfType("*types.Profile")).Return(expectedProfile, nil)

	// Execute
	resp, err := app.Test(req)

	// Assertions
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read raw body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	bodyStr := string(bodyBytes)

	assert.Equal(t, 200, resp.StatusCode, "Status code mismatch. Body: "+bodyStr)

	var respBody map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &respBody); err != nil {
		t.Fatalf("Failed to decode JSON: %v. Body: %s", err, bodyStr)
	}

	if val, ok := respBody["profile"]; !ok {
		t.Fatalf("Response did not contain 'profile'. Body: %v", respBody)
	} else if val == nil {
		t.Fatalf("Profile is nil. Body: %v", respBody)
	}

	profileMap := respBody["profile"].(map[string]interface{})
	assert.Equal(t, "My Bio", profileMap["bio"])

	mockProfileSvc.AssertExpectations(t)
}

// DummyUserService satisfies types.IUserService
type DummyUserService struct{}

func (d *DummyUserService) Save(user *types.User) (*types.User, error)              { return nil, nil }
func (d *DummyUserService) ViewByID(id uint, user *types.User) (*types.User, error) { return nil, nil }
func (d *DummyUserService) ViewByEmail(email string, user *types.User) (*types.User, error) {
	return nil, nil
}
func (d *DummyUserService) ViewVerifiedUserByEmail(email string, user *types.User) (*types.User, error) {
	return nil, nil
}
func (d *DummyUserService) List(filters *types.UserFilters) (*types.UserPage, error) { return nil, nil }
func (d *DummyUserService) Modify(id uint, updatedValues map[string]interface{}) (*types.User, error) {
	return nil, nil
}

type DummyAuthService struct{}

func (d *DummyAuthService) SignUp(userCreate *auth.UserCreateRequest) (*model.User, string, time.Time, error) {
	return nil, "", time.Time{}, nil
}
func (d *DummyAuthService) SendActivationEmail(email string) (*model.User, string, time.Time, error) {
	return nil, "", time.Time{}, nil
}
func (d *DummyAuthService) ActivateAccount(token string) (*model.User, error) { return nil, nil }
func (d *DummyAuthService) Login(username string, password string, userAgent *string, ipAddress *string) (*auth.UserResponse, error) {
	return nil, nil
}
func (d *DummyAuthService) Login2FA(tempToken string, totpToken string, userAgent *string, ipAddress *string) (*auth.UserResponse, error) {
	return nil, nil
}
func (d *DummyAuthService) GetAccessToken(refreshToken string) (*auth.TokenModel, error) {
	return nil, nil
}
func (d *DummyAuthService) RequestPasswordReset(email string) (*model.User, string, time.Time, error) {
	return nil, "", time.Time{}, nil
}
func (d *DummyAuthService) ResetPassword(token string, email string, password string) (*model.User, error) {
	return nil, nil
}
func (d *DummyAuthService) Logout(refreshToken string) error                  { return nil }
func (d *DummyAuthService) LogoutAll(userID uint) error                       { return nil }
func (d *DummyAuthService) ListSessions(userID uint) ([]model.Session, error) { return nil, nil }
func (d *DummyAuthService) ChangeEmail(username string, newEmail string, password string) error {
	return nil
}
func (d *DummyAuthService) AdminRevokeSessions(userID uint) error             { return nil }
func (d *DummyAuthService) Enable2FA(username string) (string, string, error) { return "", "", nil }
func (d *DummyAuthService) Disable2FA(username string) error                  { return nil }
func (d *DummyAuthService) ValidateAccessToken(token string) (*auth.AuthClaims, error) {
	return &auth.AuthClaims{
		UserID:    1,
		Username:  "testuser",
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}, nil
}
