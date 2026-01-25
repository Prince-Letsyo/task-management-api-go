package auth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/pkg"
	"github.com/Prince-Letsyo/task-management-api-go/internal/types"
)

// MockUserService is a mock implementation of types.IUserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Save(user *types.User) (*types.User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockUserService) ViewByID(id uint, user *types.User) (*types.User, error) {
	args := m.Called(id, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockUserService) ViewByEmail(email string, user *types.User) (*types.User, error) {
	args := m.Called(email, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockUserService) ViewVerifiedUserByEmail(email string, user *types.User) (*types.User, error) {
	args := m.Called(email, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockUserService) List(filters *types.UserFilters) (*types.UserPage, error) {
	args := m.Called(filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.UserPage), args.Error(1)
}

func (m *MockUserService) Modify(id uint, updatedValues map[string]interface{}) (*types.User, error) {
	args := m.Called(id, updatedValues)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}

func TestRegisterService_NewUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockUserSvc := new(MockUserService)
		service, _ := newRegisterService(withDatabaseRegisterRepository(mockUserSvc))

		registerForm := &types.RegisterForm{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			UserName:  "johndoe",
			Password:  "password123",
		}

		// Expect ViewByEmail to NOT find a user (return error or nil user)
		// The service implementation ignores the error from ViewByEmail and checks if user is nil.
		// Wait, look at service.go:
		// if user, _ := registerService.ViewByEmail(checkUser.Email, &checkUser); user != nil { return error }
		// So we mock ViewByEmail to return nil, nil (or nil, error)
		mockUserSvc.On("ViewByEmail", registerForm.Email, mock.AnythingOfType("*types.User")).Return(nil, errors.New("not found"))

		// Expect Save to be called
		mockUserSvc.On("Save", mock.MatchedBy(func(u *types.User) bool {
			return u.Email == registerForm.Email && u.FirstName == registerForm.FirstName
		})).Return(&types.User{ID: 1}, nil)

		user, err := service.NewUser(registerForm)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		mockUserSvc.AssertExpectations(t)
	})

	t.Run("UserAlreadyExists", func(t *testing.T) {
		mockUserSvc := new(MockUserService)
		service, _ := newRegisterService(withDatabaseRegisterRepository(mockUserSvc))

		registerForm := &types.RegisterForm{
			Email: "exist@example.com",
		}

		// Expect ViewByEmail to find a user
		mockUserSvc.On("ViewByEmail", registerForm.Email, mock.AnythingOfType("*types.User")).Return(&types.User{ID: 1}, nil)

		user, err := service.NewUser(registerForm)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "user already exists", err.Error())
		mockUserSvc.AssertExpectations(t)
	})
}

func TestLoginService_CheckLogin(t *testing.T) {
	// Setup password hasher just in case (though default is usually fine)
	// We'll rely on pkg.DefaultPassword being functional.

	t.Run("Success", func(t *testing.T) {
		mockUserSvc := new(MockUserService)
		appCfg := &config.AppCfg{}
		service, _ := newLoginService(appCfg, withDatabaseLoginRepository(mockUserSvc))

		password := "secret123"
		hashedPassword, _ := pkg.DefaultPassword.Create(password)

		loginForm := &types.Login{
			Email:    "login@example.com",
			Password: password,
		}

		validUser := &types.User{
			ID:         1,
			Email:      loginForm.Email,
			Password:   hashedPassword,
			IsVerified: true,
		}

		mockUserSvc.On("ViewVerifiedUserByEmail", loginForm.Email, mock.AnythingOfType("*types.User")).Return(validUser, nil)

		user, err := service.CheckLogin(loginForm)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, validUser.Email, user.Email)
		mockUserSvc.AssertExpectations(t)
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		mockUserSvc := new(MockUserService)
		appCfg := &config.AppCfg{}
		service, _ := newLoginService(appCfg, withDatabaseLoginRepository(mockUserSvc))

		password := "secret123"
		hashedPassword, _ := pkg.DefaultPassword.Create(password)

		loginForm := &types.Login{
			Email:    "login@example.com",
			Password: "wrongpassword",
		}

		validUser := &types.User{
			ID:         1,
			Email:      loginForm.Email,
			Password:   hashedPassword,
			IsVerified: true,
		}

		mockUserSvc.On("ViewVerifiedUserByEmail", loginForm.Email, mock.AnythingOfType("*types.User")).Return(validUser, nil)

		user, err := service.CheckLogin(loginForm)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "invalid Username or Password")
		mockUserSvc.AssertExpectations(t)
	})

	t.Run("UserNotVerified", func(t *testing.T) {
		mockUserSvc := new(MockUserService)
		appCfg := &config.AppCfg{}
		service, _ := newLoginService(appCfg, withDatabaseLoginRepository(mockUserSvc))

		loginForm := &types.Login{
			Email: "unverified@example.com",
		}

		unverifiedUser := &types.User{
			ID:         1,
			Email:      loginForm.Email,
			IsVerified: false,
		}

		// ViewVerifiedUserByEmail usually returns only verified users, but based on the code in service.go:95
		// it checks user.IsVerified manually AFTER fetching?
		// Wait, let's check service.go again.
		// user, errUser := loginService.ViewVerifiedUserByEmail(login.Email, user)
		// if errUser != nil { return nil, errUser }
		// else if !user.IsVerified { return user, error }
		//
		// This implies ViewVerifiedUserByEmail MIGHT return an unverified user?
		// Or maybe the method name is misleading?
		// Assuming ViewVerifiedUserByEmail fetches the user by email, likely filtered by some logic.
		// But the service logic explicitly checks `!user.IsVerified` on line 95.
		// So we Mock it to return the user.

		mockUserSvc.On("ViewVerifiedUserByEmail", loginForm.Email, mock.AnythingOfType("*types.User")).Return(unverifiedUser, nil)

		user, err := service.CheckLogin(loginForm)
		assert.Error(t, err)
		assert.NotNil(t, user) // It returns the user object in the error case too? Yes, line 96: return user, errors.New(...)
		assert.Contains(t, err.Error(), "account is not verified")
		mockUserSvc.AssertExpectations(t)
	})
}
