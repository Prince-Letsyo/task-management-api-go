package profile

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Prince-Letsyo/task-management-api-go/internal/types"
)

// MockProfileRepository is a mock implementation of IProfileRepository
type MockProfileRepository struct {
	mock.Mock
}

func (m *MockProfileRepository) retrieve(userID uint, profile *types.Profile) (*types.Profile, error) {
	args := m.Called(userID, profile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Profile), args.Error(1)
}

func (m *MockProfileRepository) update(id uint, profile *types.Profile) (*types.Profile, error) {
	args := m.Called(id, profile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Profile), args.Error(1)
}

func TestProfileService_View(t *testing.T) {
	mockRepo := new(MockProfileRepository)
	service, _ := newProfileService(withDatabaseProfileRepository(mockRepo))

	userID := uint(1)
	bio := "Test Bio"
	expectedProfile := &types.Profile{UserID: userID, Bio: &bio}

	mockRepo.On("retrieve", userID, mock.AnythingOfType("*types.Profile")).Return(expectedProfile, nil)

	profile, err := service.View(userID, &types.Profile{})

	assert.NoError(t, err)
	assert.Equal(t, *expectedProfile.Bio, *profile.Bio)
	mockRepo.AssertExpectations(t)
}

func TestProfileService_Modify(t *testing.T) {
	mockRepo := new(MockProfileRepository)
	service, _ := newProfileService(withDatabaseProfileRepository(mockRepo))

	userID := uint(1)
	bio := "Updated Bio"
	updates := &types.Profile{Bio: &bio}
	updatedProfile := &types.Profile{UserID: userID, Bio: &bio}

	mockRepo.On("update", userID, updates).Return(updatedProfile, nil)

	profile, err := service.Modify(userID, updates)

	assert.NoError(t, err)
	assert.Equal(t, "Updated Bio", *profile.Bio)
	mockRepo.AssertExpectations(t)
}
