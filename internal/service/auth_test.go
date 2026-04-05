package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/ronaldocristover/lms-backend/internal/model"
)

// MockUserRepository and MockRoleRepository are in user_test.go

func newTestAuthService() (AuthService, *MockUserRepository, *MockRoleRepository) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	return NewAuthService(mockUserRepo, mockRoleRepo, "test-secret-min-32-characters-long!!", 15*time.Minute, 7*24*time.Hour), mockUserRepo, mockRoleRepo
}

func TestAuthService_Register_Success(t *testing.T) {
	authSvc, mockUserRepo, mockRoleRepo := newTestAuthService()
	roleID := uuid.New()
	role := &model.Role{ID: roleID, Name: model.RoleStudent}
	req := &model.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		RoleID:   roleID,
	}
	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(role, nil)
	mockUserRepo.On("GetByEmail", mock.Anything, req.Email).Return(nil, nil)
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)
	resp, err := authSvc.Register(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, req.Email, resp.User.Email)
	assert.Equal(t, req.Name, resp.User.Name)
	assert.NotEmpty(t, resp.Token)
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}
func TestAuthService_Register_UserExists(t *testing.T) {
	authSvc, mockUserRepo, _ := newTestAuthService()
	req := &model.RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}
	existingUser := &model.User{
		ID:    uuid.New(),
		Email: req.Email,
	}
	mockUserRepo.On("GetByEmail", mock.Anything, req.Email).Return(existingUser, nil)
	resp, err := authSvc.Register(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, ErrUserExists, err)
	assert.Nil(t, resp)
	mockUserRepo.AssertExpectations(t)
}
func TestAuthService_Register_InvalidRole(t *testing.T) {
	authSvc, mockUserRepo, mockRoleRepo := newTestAuthService()
	roleID := uuid.New()
	req := &model.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		RoleID:   roleID,
	}
	mockUserRepo.On("GetByEmail", mock.Anything, req.Email).Return(nil, nil)
	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(nil, assert.AnError)
	resp, err := authSvc.Register(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}
func TestAuthService_Login_Success(t *testing.T) {
	authSvc, mockUserRepo, _ := newTestAuthService()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	roleID := uuid.New()
	user := &model.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		Name:         "Test User",
		RoleID:       roleID,
		Role:         &model.Role{ID: roleID, Name: model.RoleStudent},
	}
	mockUserRepo.On("GetByEmail", mock.Anything, user.Email).Return(user, nil)
	resp, err := authSvc.Login(context.Background(), &model.LoginRequest{
		Email:    user.Email,
		Password: "password123",
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, user.Email, resp.User.Email)
	assert.NotEmpty(t, resp.Token)
	mockUserRepo.AssertExpectations(t)
}
func TestAuthService_Login_WrongPassword(t *testing.T) {
	authSvc, mockUserRepo, _ := newTestAuthService()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	roleID := uuid.New()
	user := &model.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		RoleID:       roleID,
		Role:         &model.Role{ID: roleID, Name: model.RoleStudent},
	}
	mockUserRepo.On("GetByEmail", mock.Anything, user.Email).Return(user, nil)
	resp, err := authSvc.Login(context.Background(), &model.LoginRequest{
		Email:    user.Email,
		Password: "wrongpassword",
	})
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, resp)
	mockUserRepo.AssertExpectations(t)
}
func TestAuthService_Login_UserNotFound(t *testing.T) {
	authSvc, mockUserRepo, _ := newTestAuthService()
	mockUserRepo.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, assert.AnError)
	resp, err := authSvc.Login(context.Background(), &model.LoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	})
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, resp)
	mockUserRepo.AssertExpectations(t)
}
func TestAuthService_RefreshToken_Success(t *testing.T) {
	authSvc, mockUserRepo, _ := newTestAuthService()
	userID := uuid.New()
	roleID := uuid.New()
	user := &model.User{ID: userID, Email: "test@example.com", RoleID: roleID, Role: &model.Role{ID: roleID, Name: "student"}}
	// Generate a valid refresh token using the service internals
	svcImpl := authSvc.(*authService).UserSvc
	refreshToken, err := svcImpl.generateToken(user, "refresh", svcImpl.refreshExpiry)
	assert.NoError(t, err)
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	resp, err := authSvc.RefreshToken(context.Background(), refreshToken)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, userID, resp.User.ID)
	mockUserRepo.AssertExpectations(t)
}
func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	authSvc, _, _ := newTestAuthService()
	resp, err := authSvc.RefreshToken(context.Background(), "invalid-token")
	assert.Nil(t, resp)
	assert.Equal(t, ErrInvalidToken, err)
}
func TestAuthService_RefreshToken_UsesAccessToken(t *testing.T) {
	authSvc, mockUserRepo, _ := newTestAuthService()
	userID := uuid.New()
	roleID := uuid.New()
	user := &model.User{ID: userID, Email: "test@example.com", RoleID: roleID, Role: &model.Role{ID: roleID, Name: "student"}}
	// Generate an ACCESS token, not refresh
	svcImpl := authSvc.(*authService).UserSvc
	accessToken, err := svcImpl.generateToken(user, "access", svcImpl.jwtExpiry)
	assert.NoError(t, err)
	resp, err := authSvc.RefreshToken(context.Background(), accessToken)
	assert.Nil(t, resp)
	assert.Equal(t, ErrInvalidToken, err)
	mockUserRepo.AssertNotCalled(t, "GetByID")
}
func TestAuthService_RefreshToken_UserNotFound(t *testing.T) {
	authSvc, mockUserRepo, _ := newTestAuthService()
	userID := uuid.New()
	user := &model.User{ID: userID, Email: "gone@example.com"}
	svcImpl := authSvc.(*authService).UserSvc
	refreshToken, err := svcImpl.generateToken(user, "refresh", svcImpl.refreshExpiry)
	assert.NoError(t, err)
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(nil, assert.AnError)
	resp, err := authSvc.RefreshToken(context.Background(), refreshToken)
	assert.Nil(t, resp)
	assert.Equal(t, ErrUserNotFound, err)
	mockUserRepo.AssertExpectations(t)
}
func TestAuthService_Login_ReturnsRefreshToken(t *testing.T) {
	authSvc, mockUserRepo, _ := newTestAuthService()
	roleID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &model.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		RoleID:       roleID,
		Role:         &model.Role{ID: roleID, Name: "student"},
	}
	mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
	resp, err := authSvc.Login(context.Background(), &model.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.RefreshToken)
	mockUserRepo.AssertExpectations(t)
}
func TestAuthService_Register_ReturnsRefreshToken(t *testing.T) {
	authSvc, mockUserRepo, mockRoleRepo := newTestAuthService()
	roleID := uuid.New()
	role := &model.Role{ID: roleID, Name: model.RoleStudent}
	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(role, nil)
	mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, nil)
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)
	resp, err := authSvc.Register(context.Background(), &model.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		RoleID:   roleID,
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.RefreshToken)
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}