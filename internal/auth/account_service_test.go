package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/Prince-Letsyo/task-management-api-go/internal/model"
)

func TestAccountService_SignUp(t *testing.T) {
	auth, jwt, _ := authTestConfig()
	repo := new(MockAuthRepository)
	service := NewAccountService(repo, auth, jwt)

	req := &UserCreateRequest{
		FirstName: "Alice",
		LastName:  "Smith",
		UserName:  "alice",
		Email:     "alice@example.com",
		Password:  "Password123!",
	}

	createdUser := &model.User{
		Model:     gorm.Model{ID: 1},
		FirstName: req.FirstName,
		LastName:  req.LastName,
		UserName:  req.UserName,
		Email:     req.Email,
	}

	repo.On("CreateUser", mock.Anything).Return(createdUser, nil)

	user, token, expires, err := service.SignUp(req)
	assert.NoError(t, err)
	assert.Equal(t, createdUser.ID, user.ID)
	assert.NotEmpty(t, token)
	assert.True(t, expires.After(time.Now()))

	claims, err := parseToken(jwt.Secret, token)
	assert.NoError(t, err)
	assert.Equal(t, "activate", claims.TokenType)
	assert.Equal(t, "alice", claims.Username)

	repo.AssertExpectations(t)
}

func TestAccountService_ActivateAccount(t *testing.T) {
	auth, jwt, _ := authTestConfig()
	repo := new(MockAuthRepository)
	service := NewAccountService(repo, auth, jwt)

	claims := buildClaims(AuthClaims{
		Username:  "alice",
		TokenType: "activate",
	}, time.Minute)
	token, _ := createToken(jwt.Secret, claims)

	repo.On("ActivateUserAccount", "alice").Return(&model.User{UserName: "alice", IsVerified: true}, nil)

	user, err := service.ActivateAccount(token)
	assert.NoError(t, err)
	assert.True(t, user.IsVerified)

	repo.AssertExpectations(t)
}
