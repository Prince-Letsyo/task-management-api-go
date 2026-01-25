package service

import (
	"encoding/json"
	"net/http"
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

// MockUserService is a mock implementation for tests
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Save(user *types.User) (*types.User, error)              { return nil, nil }
func (m *MockUserService) ViewByID(id uint, user *types.User) (*types.User, error) { return nil, nil }
func (m *MockUserService) ViewByEmail(email string, user *types.User) (*types.User, error) {
	args := m.Called(email, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}
func (m *MockUserService) ViewVerifiedUserByEmail(email string, user *types.User) (*types.User, error) {
	return nil, nil
}
func (m *MockUserService) List(filters *types.UserFilters) (*types.UserPage, error) { return nil, nil }
func (m *MockUserService) Modify(id uint, updatedValues map[string]interface{}) (*types.User, error) {
	return nil, nil
}

func TestAccountService_Login(t *testing.T) {
	app := fiber.New()
	store := session.New()
	appCfg := &config.AppCfg{
		Session: config.SessionConfig{Store: store},
		JwtSecrets: config.JwtSecrets{
			Secret:        "test",
			AccessExpire:  time.Minute,
			RefreshExpire: time.Hour,
		},
	}
	mockUserSvc := new(MockUserService)
	svc := NewAccountService(appCfg, mockUserSvc)

	app.Post("/login", func(c *fiber.Ctx) error {
		user := &types.User{ID: 1, Email: "test@test.com"}
		return svc.Login(c, user)
	})

	req := httptest.NewRequest("POST", "/login", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 200, resp.StatusCode)

	// Verify Cookies
	cookies := resp.Cookies()
	hasAccess := false
	hasRefresh := false
	for _, c := range cookies {
		if c.Name == "access-token" {
			hasAccess = true
		}
		if c.Name == "refresh-token" {
			hasRefresh = true
		}
	}
	assert.True(t, hasAccess)
	assert.True(t, hasRefresh)
}

func TestAccountService_UserID(t *testing.T) {
	app := fiber.New()
	appCfg := &config.AppCfg{
		JwtSecrets: config.JwtSecrets{Secret: "test"},
	}
	mockUserSvc := new(MockUserService)
	svc := NewAccountService(appCfg, mockUserSvc)

	app.Get("/user-id", func(c *fiber.Ctx) error {
		id, err := svc.UserID(c)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.JSON(fiber.Map{"id": id})
	})

	user := types.User{ID: 5, Email: "test@test.com"}
	claims := config.NewUserClaims(user, time.Minute)
	token, _ := appCfg.JwtSecrets.NewAccessToken(claims)

	req := httptest.NewRequest("GET", "/user-id", nil)
	req.AddCookie(&http.Cookie{Name: "access-token", Value: token})

	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, float64(5), result["id"])
}
