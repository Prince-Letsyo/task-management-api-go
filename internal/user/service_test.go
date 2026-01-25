package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Prince-Letsyo/task-management-api-go/config"
	"github.com/Prince-Letsyo/task-management-api-go/internal/types"
)

// MockUserRepository is a mock implementation of IUserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) store(user *types.User) (*types.User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockUserRepository) retrieveByID(id uint, user *types.User) (*types.User, error) {
	args := m.Called(id, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockUserRepository) retrieveByEmail(email string, user *types.User) (*types.User, error) {
	args := m.Called(email, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockUserRepository) retrieveVerifiedUserByEmail(email string, user *types.User) (*types.User, error) {
	args := m.Called(email, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockUserRepository) retrievePage(filters *types.UserFilters) (*types.UserPage, error) {
	args := m.Called(filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.UserPage), args.Error(1)
}

func (m *MockUserRepository) update(id uint, updatedValues map[string]interface{}) (*types.User, error) {
	args := m.Called(id, updatedValues)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}

func TestUserService_Save(t *testing.T) {
	mockRepo := new(MockUserRepository)
	appCfg := &config.AppCfg{}
	service, _ := NewUserService(appCfg)
	service.IUserRepository = mockRepo

	user := &types.User{
		FirstName: "John",
		Email:     "john@example.com",
	}

	mockRepo.On("store", user).Return(&types.User{ID: 1, Email: "john@example.com"}, nil)

	createdUser, err := service.Save(user)

	assert.NoError(t, err)
	assert.NotNil(t, createdUser)
	assert.Equal(t, uint(1), createdUser.ID)
	mockRepo.AssertExpectations(t)
}

func TestUserService_ViewByEmail(t *testing.T) {
	mockRepo := new(MockUserRepository)
	appCfg := &config.AppCfg{}
	service, _ := NewUserService(appCfg)
	service.IUserRepository = mockRepo

	email := "john@example.com"
	expectedUser := &types.User{ID: 1, Email: email}

	mockRepo.On("retrieveByEmail", email, mock.AnythingOfType("*types.User")).Return(expectedUser, nil)

	user, err := service.ViewByEmail(email, &types.User{})

	assert.NoError(t, err)
	assert.Equal(t, expectedUser.ID, user.ID)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Modify(t *testing.T) {
	mockRepo := new(MockUserRepository)
	appCfg := &config.AppCfg{}
	service, _ := NewUserService(appCfg)
	service.IUserRepository = mockRepo

	id := uint(1)
	updates := map[string]any{"name": "New Name"}
	updatedUser := &types.User{ID: id, FirstName: "New Name"}

	mockRepo.On("update", id, updates).Return(updatedUser, nil)

	user, err := service.Modify(id, updates)

	assert.NoError(t, err)
	assert.Equal(t, "New Name", user.FirstName)
	mockRepo.AssertExpectations(t)
}
