package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yourusername/lms/internal/model"
	"github.com/yourusername/lms/internal/service"
)

type MockRoleService struct {
	mock.Mock
}

func (m *MockRoleService) Create(ctx context.Context, req *model.CreateRoleRequest) (*model.Role, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRoleService) GetByID(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRoleService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateRoleRequest) (*model.Role, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRoleService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoleService) List(ctx context.Context, req *model.ListRolesRequest) ([]*model.Role, int64, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Role), args.Get(1).(int64), args.Error(2)
}

func TestRoleHandler_Create_Success(t *testing.T) {
	mockSvc := new(MockRoleService)
	h := NewRoleHandler(mockSvc)
	router := setupRoleRouter(h)

	roleID := uuid.New()
	created := &model.Role{ID: roleID, Name: model.RoleStudent}

	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateRoleRequest")).Return(created, nil)

	body, _ := json.Marshal(model.CreateRoleRequest{Name: model.RoleStudent})
	req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestRoleHandler_Create_Duplicate(t *testing.T) {
	mockSvc := new(MockRoleService)
	h := NewRoleHandler(mockSvc)
	router := setupRoleRouter(h)

	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateRoleRequest")).Return(nil, service.ErrRoleExists)

	body, _ := json.Marshal(model.CreateRoleRequest{Name: model.RoleAdmin})
	req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestRoleHandler_Create_InvalidRole(t *testing.T) {
	mockSvc := new(MockRoleService)
	h := NewRoleHandler(mockSvc)
	router := setupRoleRouter(h)

	body, _ := json.Marshal(map[string]string{"name": "invalid_role"})
	req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRoleHandler_Get_Success(t *testing.T) {
	mockSvc := new(MockRoleService)
	h := NewRoleHandler(mockSvc)
	router := setupRoleRouter(h)

	roleID := uuid.New()
	role := &model.Role{ID: roleID, Name: model.RoleAdmin}

	mockSvc.On("GetByID", mock.Anything, roleID).Return(role, nil)

	req := httptest.NewRequest(http.MethodGet, "/roles/"+roleID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestRoleHandler_Get_InvalidID(t *testing.T) {
	mockSvc := new(MockRoleService)
	h := NewRoleHandler(mockSvc)
	router := setupRoleRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/roles/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRoleHandler_Get_NotFound(t *testing.T) {
	mockSvc := new(MockRoleService)
	h := NewRoleHandler(mockSvc)
	router := setupRoleRouter(h)

	roleID := uuid.New()
	mockSvc.On("GetByID", mock.Anything, roleID).Return(nil, service.ErrRoleNotFound)

	req := httptest.NewRequest(http.MethodGet, "/roles/"+roleID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestRoleHandler_Update_Success(t *testing.T) {
	mockSvc := new(MockRoleService)
	h := NewRoleHandler(mockSvc)
	router := setupRoleRouter(h)

	roleID := uuid.New()
	updated := &model.Role{ID: roleID, Name: model.RoleTutor}

	mockSvc.On("Update", mock.Anything, roleID, mock.AnythingOfType("*model.UpdateRoleRequest")).Return(updated, nil)

	body, _ := json.Marshal(model.UpdateRoleRequest{Name: model.RoleTutor})
	req := httptest.NewRequest(http.MethodPut, "/roles/"+roleID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestRoleHandler_Update_InvalidID(t *testing.T) {
	mockSvc := new(MockRoleService)
	h := NewRoleHandler(mockSvc)
	router := setupRoleRouter(h)

	body, _ := json.Marshal(model.UpdateRoleRequest{Name: model.RoleAdmin})
	req := httptest.NewRequest(http.MethodPut, "/roles/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRoleHandler_Update_NotFound(t *testing.T) {
	mockSvc := new(MockRoleService)
	h := NewRoleHandler(mockSvc)
	router := setupRoleRouter(h)

	roleID := uuid.New()
	mockSvc.On("Update", mock.Anything, roleID, mock.AnythingOfType("*model.UpdateRoleRequest")).Return(nil, service.ErrRoleNotFound)

	body, _ := json.Marshal(model.UpdateRoleRequest{Name: model.RoleAdmin})
	req := httptest.NewRequest(http.MethodPut, "/roles/"+roleID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestRoleHandler_Delete_Success(t *testing.T) {
	mockSvc := new(MockRoleService)
	h := NewRoleHandler(mockSvc)
	router := setupRoleRouter(h)

	roleID := uuid.New()
	mockSvc.On("Delete", mock.Anything, roleID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/roles/"+roleID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestRoleHandler_Delete_InvalidID(t *testing.T) {
	mockSvc := new(MockRoleService)
	h := NewRoleHandler(mockSvc)
	router := setupRoleRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/roles/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRoleHandler_Delete_NotFound(t *testing.T) {
	mockSvc := new(MockRoleService)
	h := NewRoleHandler(mockSvc)
	router := setupRoleRouter(h)

	roleID := uuid.New()
	mockSvc.On("Delete", mock.Anything, roleID).Return(service.ErrRoleNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/roles/"+roleID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestRoleHandler_List_Success(t *testing.T) {
	mockSvc := new(MockRoleService)
	h := NewRoleHandler(mockSvc)
	router := setupRoleRouter(h)

	roles := []*model.Role{
		{ID: uuid.New(), Name: model.RoleAdmin},
		{ID: uuid.New(), Name: model.RoleStudent},
	}

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListRolesRequest) bool {
		return req.Page == 1 && req.PageSize == 20
	})).Return(roles, int64(2), nil)

	req := httptest.NewRequest(http.MethodGet, "/roles?page=1&page_size=20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestRoleHandler_List_WithSearch(t *testing.T) {
	mockSvc := new(MockRoleService)
	h := NewRoleHandler(mockSvc)
	router := setupRoleRouter(h)

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListRolesRequest) bool {
		return req.Search == "admin"
	})).Return([]*model.Role{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/roles?search=admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestRoleHandler_List_Empty(t *testing.T) {
	mockSvc := new(MockRoleService)
	h := NewRoleHandler(mockSvc)
	router := setupRoleRouter(h)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListRolesRequest")).Return([]*model.Role{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestRoleHandler_List_ServiceError(t *testing.T) {
	mockSvc := new(MockRoleService)
	h := NewRoleHandler(mockSvc)
	router := setupRoleRouter(h)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListRolesRequest")).Return(([]*model.Role)(nil), int64(0), assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func setupRoleRouter(handler *RoleHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/roles", handler.Create)
	r.GET("/roles", handler.List)
	r.GET("/roles/:id", handler.Get)
	r.PUT("/roles/:id", handler.Update)
	r.DELETE("/roles/:id", handler.Delete)
	return r
}
