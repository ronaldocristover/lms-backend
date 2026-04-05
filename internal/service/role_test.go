package service

import (
	"context"
	"testing"
	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ronaldocristover/lms-backend/internal/model"
)

type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) Create(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	if args.Error(0) == nil {
		role.ID = uuid.New()
	}
	return args.Error(0)
}

func (m *MockRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRoleRepository) GetByName(ctx context.Context, name string) (*model.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRoleRepository) Update(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoleRepository) List(ctx context.Context, filter *model.ListRolesRequest) ([]*model.Role, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Role), args.Get(1).(int64), args.Error(2)
}

func newTestRoleService() (RoleService, *MockRoleRepository) {
	mockRepo := new(MockRoleRepository)
	return NewRoleService(mockRepo, zap.NewNop().Sugar()), mockRepo
}

func TestRoleService_Create_Success(t *testing.T) {
	svc, mockRepo := newTestRoleService()

	req := &model.CreateRoleRequest{
		Name: model.RoleStudent,
	}

	mockRepo.On("GetByName", mock.Anything, req.Name).Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Role")).Return(nil)

	role, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, req.Name, role.Name)
	mockRepo.AssertExpectations(t)
}

func TestRoleService_Create_RoleExists(t *testing.T) {
	svc, mockRepo := newTestRoleService()

	req := &model.CreateRoleRequest{
		Name: model.RoleAdmin,
	}

	existing := &model.Role{ID: uuid.New(), Name: model.RoleAdmin}
	mockRepo.On("GetByName", mock.Anything, req.Name).Return(existing, nil)

	role, err := svc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, ErrRoleExists, err)
	assert.Nil(t, role)
	mockRepo.AssertExpectations(t)
}

func TestRoleService_GetByID_Success(t *testing.T) {
	svc, mockRepo := newTestRoleService()

	roleID := uuid.New()
	expected := &model.Role{ID: roleID, Name: model.RoleAdmin}

	mockRepo.On("GetByID", mock.Anything, roleID).Return(expected, nil)

	role, err := svc.GetByID(context.Background(), roleID)

	assert.NoError(t, err)
	assert.Equal(t, expected, role)
	mockRepo.AssertExpectations(t)
}

func TestRoleService_GetByID_NotFound(t *testing.T) {
	svc, mockRepo := newTestRoleService()

	roleID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, roleID).Return(nil, assert.AnError)

	role, err := svc.GetByID(context.Background(), roleID)

	assert.Equal(t, ErrRoleNotFound, err)
	assert.Nil(t, role)
	mockRepo.AssertExpectations(t)
}

func TestRoleService_Update_Success(t *testing.T) {
	svc, mockRepo := newTestRoleService()

	roleID := uuid.New()
	existing := &model.Role{ID: roleID, Name: model.RoleStudent}

	mockRepo.On("GetByID", mock.Anything, roleID).Return(existing, nil)
	mockRepo.On("GetByName", mock.Anything, model.RoleTutor).Return(nil, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Role")).Return(nil)

	role, err := svc.Update(context.Background(), roleID, &model.UpdateRoleRequest{
		Name: model.RoleTutor,
	})

	assert.NoError(t, err)
	assert.Equal(t, model.RoleTutor, role.Name)
	mockRepo.AssertExpectations(t)
}

func TestRoleService_Update_NotFound(t *testing.T) {
	svc, mockRepo := newTestRoleService()

	roleID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, roleID).Return(nil, assert.AnError)

	role, err := svc.Update(context.Background(), roleID, &model.UpdateRoleRequest{Name: model.RoleAdmin})

	assert.Equal(t, ErrRoleNotFound, err)
	assert.Nil(t, role)
	mockRepo.AssertExpectations(t)
}

func TestRoleService_Update_DuplicateName(t *testing.T) {
	svc, mockRepo := newTestRoleService()

	roleID := uuid.New()
	otherID := uuid.New()
	existing := &model.Role{ID: roleID, Name: model.RoleStudent}
	duplicate := &model.Role{ID: otherID, Name: model.RoleAdmin}

	mockRepo.On("GetByID", mock.Anything, roleID).Return(existing, nil)
	mockRepo.On("GetByName", mock.Anything, model.RoleAdmin).Return(duplicate, nil)

	role, err := svc.Update(context.Background(), roleID, &model.UpdateRoleRequest{Name: model.RoleAdmin})

	assert.Equal(t, ErrRoleExists, err)
	assert.Nil(t, role)
	mockRepo.AssertExpectations(t)
}

func TestRoleService_Delete_Success(t *testing.T) {
	svc, mockRepo := newTestRoleService()

	roleID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, roleID).Return(&model.Role{ID: roleID}, nil)
	mockRepo.On("Delete", mock.Anything, roleID).Return(nil)

	err := svc.Delete(context.Background(), roleID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRoleService_Delete_NotFound(t *testing.T) {
	svc, mockRepo := newTestRoleService()

	roleID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, roleID).Return(nil, assert.AnError)

	err := svc.Delete(context.Background(), roleID)

	assert.Equal(t, ErrRoleNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestRoleService_List_Success(t *testing.T) {
	svc, mockRepo := newTestRoleService()

	roles := []*model.Role{
		{ID: uuid.New(), Name: model.RoleAdmin},
		{ID: uuid.New(), Name: model.RoleStudent},
	}

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListRolesRequest")).Return(roles, int64(2), nil)

	result, total, err := svc.List(context.Background(), &model.ListRolesRequest{Page: 1, PageSize: 20})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestRoleService_List_DefaultPagination(t *testing.T) {
	svc, mockRepo := newTestRoleService()

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListRolesRequest")).Return([]*model.Role{}, int64(0), nil)

	result, total, err := svc.List(context.Background(), &model.ListRolesRequest{})

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}

func TestRoleService_List_WithSearch(t *testing.T) {
	svc, mockRepo := newTestRoleService()

	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListRolesRequest) bool {
		return req.Search == "admin"
	})).Return([]*model.Role{}, int64(0), nil)

	result, _, err := svc.List(context.Background(), &model.ListRolesRequest{
		Page:     1,
		PageSize: 20,
		Search:   "admin",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRepo.AssertExpectations(t)
}
