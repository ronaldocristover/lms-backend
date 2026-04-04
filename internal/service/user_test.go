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

func (m *MockUserRepository) List(ctx context.Context, filter *model.ListUsersRequest) ([]*model.User, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.User), args.Get(1).(int64), args.Error(2)
}

func newTestService() (UserService, *MockUserRepository) {
	mockRepo := new(MockUserRepository)
	return NewUserService(mockRepo, "test-secret-min-32-characters-long!!", time.Hour), mockRepo
}

// ─── REGISTER ───

func TestUserService_Register_Success(t *testing.T) {
	svc, mockRepo := newTestService()

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
	assert.Equal(t, model.UserRoleUser, resp.User.Role)
	assert.Equal(t, model.UserStatusActive, resp.User.Status)
	assert.NotEmpty(t, resp.Token)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Register_UserExists(t *testing.T) {
	svc, mockRepo := newTestService()

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

// ─── LOGIN ───

func TestUserService_Login_Success(t *testing.T) {
	svc, mockRepo := newTestService()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	user := &model.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		Name:         "Test User",
		Role:         model.UserRoleUser,
		Status:       model.UserStatusActive,
	}

	mockRepo.On("GetByEmail", mock.Anything, user.Email).Return(user, nil)

	resp, err := svc.Login(context.Background(), &model.LoginRequest{
		Email:    user.Email,
		Password: "password123",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, user.Email, resp.User.Email)
	assert.NotEmpty(t, resp.Token)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Login_SuspendedUser(t *testing.T) {
	svc, mockRepo := newTestService()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	user := &model.User{
		ID:           uuid.New(),
		Email:        "suspended@example.com",
		PasswordHash: string(hashedPassword),
		Status:       model.UserStatusSuspended,
	}

	mockRepo.On("GetByEmail", mock.Anything, user.Email).Return(user, nil)

	resp, err := svc.Login(context.Background(), &model.LoginRequest{
		Email:    user.Email,
		Password: "password123",
	})

	assert.Error(t, err)
	assert.Equal(t, ErrUserSuspended, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Login_WrongPassword(t *testing.T) {
	svc, mockRepo := newTestService()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	user := &model.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		Status:       model.UserStatusActive,
	}

	mockRepo.On("GetByEmail", mock.Anything, user.Email).Return(user, nil)

	resp, err := svc.Login(context.Background(), &model.LoginRequest{
		Email:    user.Email,
		Password: "wrongpassword",
	})

	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	svc, mockRepo := newTestService()

	mockRepo.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, assert.AnError)

	resp, err := svc.Login(context.Background(), &model.LoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	})

	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

// ─── GET BY ID ───

func TestUserService_GetByID_Success(t *testing.T) {
	svc, mockRepo := newTestService()

	userID := uuid.New()
	expected := &model.User{ID: userID, Email: "test@example.com", Name: "Test User", Status: model.UserStatusActive}

	mockRepo.On("GetByID", mock.Anything, userID).Return(expected, nil)

	user, err := svc.GetByID(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, expected, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	svc, mockRepo := newTestService()

	userID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, userID).Return(nil, assert.AnError)

	user, err := svc.GetByID(context.Background(), userID)

	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

// ─── UPDATE ───

func TestUserService_Update_Success(t *testing.T) {
	svc, mockRepo := newTestService()

	userID := uuid.New()
	existing := &model.User{ID: userID, Email: "test@example.com", Name: "Old Name", Role: model.UserRoleUser, Status: model.UserStatusActive}

	mockRepo.On("GetByID", mock.Anything, userID).Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	user, err := svc.Update(context.Background(), userID, &model.UpdateUserRequest{
		Name:   "New Name",
		Avatar: "https://example.com/avatar.jpg",
		Status: model.UserStatusInactive,
	})

	assert.NoError(t, err)
	assert.Equal(t, "New Name", user.Name)
	assert.Equal(t, "https://example.com/avatar.jpg", user.Avatar)
	assert.Equal(t, model.UserStatusInactive, user.Status)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Update_NotFound(t *testing.T) {
	svc, mockRepo := newTestService()

	userID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, userID).Return(nil, assert.AnError)

	user, err := svc.Update(context.Background(), userID, &model.UpdateUserRequest{Name: "New"})

	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

// ─── LIST ───

func TestUserService_List_Success(t *testing.T) {
	svc, mockRepo := newTestService()

	users := []*model.User{
		{ID: uuid.New(), Email: "user1@example.com", Status: model.UserStatusActive},
		{ID: uuid.New(), Email: "user2@example.com", Status: model.UserStatusActive},
	}

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListUsersRequest")).Return(users, int64(2), nil)

	result, total, err := svc.List(context.Background(), &model.ListUsersRequest{Page: 1, PageSize: 20})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestUserService_List_DefaultPagination(t *testing.T) {
	svc, mockRepo := newTestService()

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListUsersRequest")).Return([]*model.User{}, int64(0), nil)

	result, total, err := svc.List(context.Background(), &model.ListUsersRequest{})

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}

func TestUserService_List_WithFilter(t *testing.T) {
	svc, mockRepo := newTestService()

	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListUsersRequest) bool {
		return req.Role == model.UserRoleTutor && req.Status == model.UserStatusActive
	})).Return([]*model.User{}, int64(0), nil)

	result, _, err := svc.List(context.Background(), &model.ListUsersRequest{
		Page:     1,
		PageSize: 20,
		Role:     model.UserRoleTutor,
		Status:   model.UserStatusActive,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRepo.AssertExpectations(t)
}

// ─── DELETE ───

func TestUserService_Delete_Success(t *testing.T) {
	svc, mockRepo := newTestService()

	userID := uuid.New()
	mockRepo.On("Delete", mock.Anything, userID).Return(nil)

	err := svc.Delete(context.Background(), userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
