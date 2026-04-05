package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/ronaldocristover/lms-backend/internal/model"
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

func newTestUserService() (UserService, *MockUserRepository, *MockRoleRepository) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	return NewUserService(mockUserRepo, mockRoleRepo, "test-secret-min-32-characters-long!!"), mockUserRepo, mockRoleRepo
}

func TestUserService_Register_Success(t *testing.T) {
	svc, mockUserRepo, mockRoleRepo := newTestUserService()

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

	resp, err := svc.Register(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, req.Email, resp.User.Email)
	assert.Equal(t, req.Name, resp.User.Name)
	assert.NotEmpty(t, resp.Token)
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func TestUserService_Register_UserExists(t *testing.T) {
	svc, mockUserRepo, _ := newTestUserService()

	req := &model.RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}

	existingUser := &model.User{
		ID:    uuid.New(),
		Email: req.Email,
	}

	mockUserRepo.On("GetByEmail", mock.Anything, req.Email).Return(existingUser, nil)

	resp, err := svc.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, ErrUserExists, err)
	assert.Nil(t, resp)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_Register_InvalidRole(t *testing.T) {
	svc, mockUserRepo, mockRoleRepo := newTestUserService()

	roleID := uuid.New()
	req := &model.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		RoleID:   roleID,
	}

	mockUserRepo.On("GetByEmail", mock.Anything, req.Email).Return(nil, nil)
	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(nil, assert.AnError)

	resp, err := svc.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func TestUserService_Login_Success(t *testing.T) {
	svc, mockUserRepo, _ := newTestUserService()

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

	resp, err := svc.Login(context.Background(), &model.LoginRequest{
		Email:    user.Email,
		Password: "password123",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, user.Email, resp.User.Email)
	assert.NotEmpty(t, resp.Token)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_Login_WrongPassword(t *testing.T) {
	svc, mockUserRepo, _ := newTestUserService()

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

	resp, err := svc.Login(context.Background(), &model.LoginRequest{
		Email:    user.Email,
		Password: "wrongpassword",
	})

	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, resp)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	svc, mockUserRepo, _ := newTestUserService()

	mockUserRepo.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, assert.AnError)

	resp, err := svc.Login(context.Background(), &model.LoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	})

	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, resp)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_Create_Success(t *testing.T) {
	svc, mockUserRepo, mockRoleRepo := newTestUserService()

	roleID := uuid.New()
	role := &model.Role{ID: roleID, Name: model.RoleStudent}

	req := &model.CreateUserRequest{
		Name:     "New User",
		Email:    "new@example.com",
		Password: "password123",
		RoleID:   roleID,
	}

	mockUserRepo.On("GetByEmail", mock.Anything, req.Email).Return(nil, nil)
	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(role, nil)
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	user, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, req.Email, user.Email)
	assert.Equal(t, req.Name, user.Name)
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func TestUserService_Create_UserExists(t *testing.T) {
	svc, mockUserRepo, _ := newTestUserService()

	roleID := uuid.New()
	req := &model.CreateUserRequest{
		Name:     "Existing",
		Email:    "existing@example.com",
		Password: "password123",
		RoleID:   roleID,
	}

	existing := &model.User{ID: uuid.New(), Email: req.Email}
	mockUserRepo.On("GetByEmail", mock.Anything, req.Email).Return(existing, nil)

	user, err := svc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, ErrUserExists, err)
	assert.Nil(t, user)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_GetByID_Success(t *testing.T) {
	svc, mockUserRepo, _ := newTestUserService()

	userID := uuid.New()
	expected := &model.User{ID: userID, Email: "test@example.com", Name: "Test User"}

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(expected, nil)

	user, err := svc.GetByID(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, expected, user)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	svc, mockUserRepo, _ := newTestUserService()

	userID := uuid.New()
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(nil, assert.AnError)

	user, err := svc.GetByID(context.Background(), userID)

	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, user)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_Update_Success(t *testing.T) {
	svc, mockUserRepo, mockRoleRepo := newTestUserService()

	userID := uuid.New()
	roleID := uuid.New()
	role := &model.Role{ID: roleID, Name: model.RoleTutor}
	existing := &model.User{ID: userID, Email: "test@example.com", Name: "Old Name", RoleID: roleID, Role: role}

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(existing, nil)
	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(role, nil)
	mockUserRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	user, err := svc.Update(context.Background(), userID, &model.UpdateUserRequest{
		Name:   "New Name",
		RoleID: roleID,
	})

	assert.NoError(t, err)
	assert.Equal(t, "New Name", user.Name)
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func TestUserService_Update_NotFound(t *testing.T) {
	svc, mockUserRepo, _ := newTestUserService()

	userID := uuid.New()
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(nil, assert.AnError)

	user, err := svc.Update(context.Background(), userID, &model.UpdateUserRequest{Name: "New"})

	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, user)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_Update_DuplicateEmail(t *testing.T) {
	svc, mockUserRepo, _ := newTestUserService()

	userID := uuid.New()
	otherID := uuid.New()
	existing := &model.User{ID: userID, Email: "old@example.com", Name: "Old Name"}
	duplicate := &model.User{ID: otherID, Email: "new@example.com"}

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(existing, nil)
	mockUserRepo.On("GetByEmail", mock.Anything, "new@example.com").Return(duplicate, nil)

	user, err := svc.Update(context.Background(), userID, &model.UpdateUserRequest{
		Email: "new@example.com",
	})

	assert.Equal(t, ErrUserExists, err)
	assert.Nil(t, user)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_List_Success(t *testing.T) {
	svc, mockUserRepo, _ := newTestUserService()

	roleID := uuid.New()
	users := []*model.User{
		{ID: uuid.New(), Email: "user1@example.com", RoleID: roleID, Role: &model.Role{ID: roleID, Name: model.RoleStudent}},
		{ID: uuid.New(), Email: "user2@example.com", RoleID: roleID, Role: &model.Role{ID: roleID, Name: model.RoleStudent}},
	}

	mockUserRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListUsersRequest")).Return(users, int64(2), nil)

	result, total, err := svc.List(context.Background(), &model.ListUsersRequest{Page: 1, PageSize: 20})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_List_DefaultPagination(t *testing.T) {
	svc, mockUserRepo, _ := newTestUserService()

	mockUserRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListUsersRequest")).Return([]*model.User{}, int64(0), nil)

	result, total, err := svc.List(context.Background(), &model.ListUsersRequest{})

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, int64(0), total)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_List_WithFilter(t *testing.T) {
	svc, mockUserRepo, _ := newTestUserService()

	roleID := uuid.New()
	mockUserRepo.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListUsersRequest) bool {
		return req.RoleID == roleID.String()
	})).Return([]*model.User{}, int64(0), nil)

	result, _, err := svc.List(context.Background(), &model.ListUsersRequest{
		Page:     1,
		PageSize: 20,
		RoleID:   roleID.String(),
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_Delete_Success(t *testing.T) {
	svc, mockUserRepo, _ := newTestUserService()

	userID := uuid.New()
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(&model.User{ID: userID}, nil)
	mockUserRepo.On("Delete", mock.Anything, userID).Return(nil)

	err := svc.Delete(context.Background(), userID)

	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_Delete_NotFound(t *testing.T) {
	svc, mockUserRepo, _ := newTestUserService()

	userID := uuid.New()
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(nil, assert.AnError)

	err := svc.Delete(context.Background(), userID)

	assert.Equal(t, ErrUserNotFound, err)
	mockUserRepo.AssertExpectations(t)
}
