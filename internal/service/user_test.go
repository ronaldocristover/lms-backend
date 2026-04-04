package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/yourusername/lms/internal/model"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	if args.Error(0) == nil {
		user.ID = uuid.New()
	}
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*model.User, int64, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*model.User), args.Get(1).(int64), args.Error(2)
}

func TestUserService_Register_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo, "test-secret", time.Hour)

	req := &model.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	mockRepo.On("GetByEmail", mock.Anything, req.Email).Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	resp, err := svc.Register(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, req.Email, resp.User.Email)
	assert.Equal(t, req.Name, resp.User.Name)
	assert.NotEmpty(t, resp.Token)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Register_UserExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo, "test-secret", time.Hour)

	req := &model.RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}

	existingUser := &model.User{
		ID:    uuid.New(),
		Email: req.Email,
	}

	mockRepo.On("GetByEmail", mock.Anything, req.Email).Return(existingUser, nil)

	resp, err := svc.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, ErrUserExists, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Login_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo, "test-secret", time.Hour)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	user := &model.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		Name:         "Test User",
		Role:         "user",
	}

	req := &model.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	mockRepo.On("GetByEmail", mock.Anything, req.Email).Return(user, nil)

	resp, err := svc.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, user.Email, resp.User.Email)
	assert.NotEmpty(t, resp.Token)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Login_InvalidCredentials_WrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo, "test-secret", time.Hour)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	user := &model.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
	}

	req := &model.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	mockRepo.On("GetByEmail", mock.Anything, req.Email).Return(user, nil)

	resp, err := svc.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Login_InvalidCredentials_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo, "test-secret", time.Hour)

	req := &model.LoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	}

	mockRepo.On("GetByEmail", mock.Anything, req.Email).Return(nil, assert.AnError)

	resp, err := svc.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetByID_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo, "test-secret", time.Hour)

	userID := uuid.New()
	expectedUser := &model.User{
		ID:    userID,
		Email: "test@example.com",
		Name:  "Test User",
	}

	mockRepo.On("GetByID", mock.Anything, userID).Return(expectedUser, nil)

	user, err := svc.GetByID(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo, "test-secret", time.Hour)

	userID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, userID).Return(nil, assert.AnError)

	user, err := svc.GetByID(context.Background(), userID)

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_List_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo, "test-secret", time.Hour)

	users := []*model.User{
		{ID: uuid.New(), Email: "user1@example.com"},
		{ID: uuid.New(), Email: "user2@example.com"},
	}

	mockRepo.On("List", mock.Anything, 20, 0).Return(users, int64(2), nil)

	result, total, err := svc.List(context.Background(), 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, users, result)
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestUserService_List_DefaultPagination(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo, "test-secret", time.Hour)

	users := []*model.User{}

	mockRepo.On("List", mock.Anything, 20, 0).Return(users, int64(0), nil)

	// page = 0 should default to 1, pageSize = 0 should default to 20
	result, total, err := svc.List(context.Background(), 0, 0)

	assert.NoError(t, err)
	assert.Equal(t, users, result)
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Delete_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo, "test-secret", time.Hour)

	userID := uuid.New()

	mockRepo.On("Delete", mock.Anything, userID).Return(nil)

	err := svc.Delete(context.Background(), userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
