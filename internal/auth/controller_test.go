package auth

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/types"
)

// MockLoginService is a mock implementation of types.ILoginService
type MockLoginService struct {
	mock.Mock
}

func (m *MockLoginService) CheckLogin(login *types.Login) (*types.User, error) {
	args := m.Called(login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}

func TestAuthController_Login(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Setup
		app := fiber.New()
		store := session.New()
		appCfg := &config.AppCfg{
			Session: config.SessionConfig{
				Store: store,
			},
			JwtSecrets: config.JwtSecrets{
				Secret:        "test-secret",
				AccessExpire:  time.Minute * 15,
				RefreshExpire: time.Hour * 24,
			},
		}

		mockUserSvc := new(MockUserService)
		mockLoginSvc := new(MockLoginService)

		authC := Auth{
			router:          app,
			userService:     mockUserSvc,
			registerService: nil,
			loginService:    mockLoginSvc,
		}

		// Initialize controller and routes
		// Note: newAuthController returns an interface but registers routes on authC.router (app)
		newAuthController(authC, appCfg)

		// Prepare Request
		loginReq := &types.Login{
			Email:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// Mock Expectations
		expectedUser := &types.User{
			ID:         1,
			Email:      "test@example.com",
			IsVerified: true,
			Password:   "hashed-password",
		}

		mockLoginSvc.On("CheckLogin", mock.MatchedBy(func(l *types.Login) bool {
			return l.Email == loginReq.Email && l.Password == loginReq.Password
		})).Return(expectedUser, nil)

		// Execute
		resp, err := app.Test(req)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		// Verify Cookies
		cookies := resp.Cookies()
		var accessCookie, refreshCookie *func(*fiber.Ctx) string // just checking existence
		_ = accessCookie
		_ = refreshCookie

		hasAccess := false
		hasRefresh := false
		for _, c := range cookies {
			if c.Name == "access-token" && c.Value != "" {
				hasAccess = true
			}
			if c.Name == "refresh-token" && c.Value != "" {
				hasRefresh = true
			}
		}
		assert.True(t, hasAccess, "access-token cookie should be set")
		assert.True(t, hasRefresh, "refresh-token cookie should be set")

		// Verify JSON Response
		var respBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&respBody)

		assert.Contains(t, respBody, "token")
		tokenMap := respBody["token"].(map[string]interface{})
		assert.NotEmpty(t, tokenMap["access"])
		assert.NotEmpty(t, tokenMap["refresh"])

		mockLoginSvc.AssertExpectations(t)
	})
}
