package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

type MockOrganizationService struct {
	mock.Mock
}

func (m *MockOrganizationService) Create(ctx context.Context, req *model.CreateOrganizationRequest) (*model.Organization, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Organization), args.Error(1)
}

func (m *MockOrganizationService) GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Organization), args.Error(1)
}

func (m *MockOrganizationService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateOrganizationRequest) (*model.Organization, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Organization), args.Error(1)
}

func (m *MockOrganizationService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrganizationService) List(ctx context.Context, req *model.ListOrganizationsRequest) ([]*model.Organization, int64, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Organization), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrganizationService) AddUser(ctx context.Context, orgID uuid.UUID, req *model.AddOrgUserRequest) (*model.OrganizationUser, error) {
	args := m.Called(ctx, orgID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OrganizationUser), args.Error(1)
}

func (m *MockOrganizationService) UpdateUserRole(ctx context.Context, orgID, orgUserID uuid.UUID, req *model.UpdateOrgUserRoleRequest) (*model.OrganizationUser, error) {
	args := m.Called(ctx, orgID, orgUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OrganizationUser), args.Error(1)
}

func (m *MockOrganizationService) RemoveUser(ctx context.Context, orgID, orgUserID uuid.UUID) error {
	args := m.Called(ctx, orgID, orgUserID)
	return args.Error(0)
}

func (m *MockOrganizationService) ListUsers(ctx context.Context, orgID uuid.UUID, req *model.ListOrgUsersRequest) ([]*model.OrganizationUser, int64, error) {
	args := m.Called(ctx, orgID, req)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.OrganizationUser), args.Get(1).(int64), args.Error(2)
}

func setupOrgRouter(handler *OrganizationHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/organizations", handler.Create)
	r.GET("/organizations", handler.List)
	r.GET("/organizations/:id", handler.Get)
	r.PUT("/organizations/:id", handler.Update)
	r.DELETE("/organizations/:id", handler.Delete)
	r.POST("/organizations/:id/users", handler.AddUser)
	r.GET("/organizations/:id/users", handler.ListUsers)
	r.PUT("/organizations/:id/users/:userId", handler.UpdateUserRole)
	r.DELETE("/organizations/:id/users/:userId", handler.RemoveUser)
	return r
}

func TestOrganizationHandler_Create_Success(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	ownerID := uuid.New()
	org := &model.Organization{
		ID:      uuid.New(),
		Name:    "Test Org",
		OwnerID: ownerID,
	}

	mockSvc.On("Create", mock.Anything, mock.MatchedBy(func(req *model.CreateOrganizationRequest) bool {
		return req.Name == "Test Org" && req.OwnerID == ownerID.String()
	})).Return(org, nil)

	body, _ := json.Marshal(model.CreateOrganizationRequest{
		Name:    "Test Org",
		OwnerID: ownerID.String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_Create_InvalidRequest(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	body, _ := json.Marshal(map[string]string{"name": ""})
	req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrganizationHandler_Create_OrgExists(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	ownerID := uuid.New()
	mockSvc.On("Create", mock.Anything, mock.Anything).Return(nil, service.ErrOrganizationExists)

	body, _ := json.Marshal(model.CreateOrganizationRequest{
		Name:    "Existing Org",
		OwnerID: ownerID.String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_List_Success(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgs := []*model.Organization{
		{ID: uuid.New(), Name: "Org 1", OwnerID: uuid.New()},
		{ID: uuid.New(), Name: "Org 2", OwnerID: uuid.New()},
	}

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListOrganizationsRequest) bool {
		return req.Page == 1 && req.PageSize == 20
	})).Return(orgs, int64(2), nil)

	req := httptest.NewRequest(http.MethodGet, "/organizations?page=1&page_size=20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_List_WithSearch(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListOrganizationsRequest) bool {
		return req.Search == "test"
	})).Return([]*model.Organization{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/organizations?search=test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_List_Empty(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListOrganizationsRequest")).Return([]*model.Organization{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_Get_Success(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	org := &model.Organization{ID: orgID, Name: "Test Org", OwnerID: uuid.New()}

	mockSvc.On("GetByID", mock.Anything, orgID).Return(org, nil)

	req := httptest.NewRequest(http.MethodGet, "/organizations/"+orgID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_Get_InvalidID(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/organizations/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrganizationHandler_Get_NotFound(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	mockSvc.On("GetByID", mock.Anything, orgID).Return(nil, service.ErrOrganizationNotFound)

	req := httptest.NewRequest(http.MethodGet, "/organizations/"+orgID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_Update_Success(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	updatedOrg := &model.Organization{ID: orgID, Name: "Updated Org", OwnerID: uuid.New()}

	mockSvc.On("Update", mock.Anything, orgID, mock.AnythingOfType("*model.UpdateOrganizationRequest")).Return(updatedOrg, nil)

	body, _ := json.Marshal(model.UpdateOrganizationRequest{Name: "Updated Org"})
	req := httptest.NewRequest(http.MethodPut, "/organizations/"+orgID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_Update_InvalidID(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	body, _ := json.Marshal(model.UpdateOrganizationRequest{Name: "Test"})
	req := httptest.NewRequest(http.MethodPut, "/organizations/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrganizationHandler_Update_NotFound(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	mockSvc.On("Update", mock.Anything, orgID, mock.Anything).Return(nil, service.ErrOrganizationNotFound)

	body, _ := json.Marshal(model.UpdateOrganizationRequest{Name: "Updated"})
	req := httptest.NewRequest(http.MethodPut, "/organizations/"+orgID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_Delete_Success(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	mockSvc.On("Delete", mock.Anything, orgID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/"+orgID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_Delete_InvalidID(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrganizationHandler_Delete_NotFound(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	mockSvc.On("Delete", mock.Anything, orgID).Return(service.ErrOrganizationNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/"+orgID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_AddUser_Success(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	userID := uuid.New()
	orgUser := &model.OrganizationUser{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		Role:           model.OrgRoleMember,
	}

	mockSvc.On("AddUser", mock.Anything, orgID, mock.AnythingOfType("*model.AddOrgUserRequest")).Return(orgUser, nil)

	body, _ := json.Marshal(model.AddOrgUserRequest{
		UserID: userID.String(),
		Role:   model.OrgRoleMember,
	})
	req := httptest.NewRequest(http.MethodPost, "/organizations/"+orgID.String()+"/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_AddUser_AlreadyInOrg(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	userID := uuid.New()

	mockSvc.On("AddUser", mock.Anything, orgID, mock.Anything).Return(nil, service.ErrUserAlreadyInOrg)

	body, _ := json.Marshal(model.AddOrgUserRequest{
		UserID: userID.String(),
		Role:   model.OrgRoleMember,
	})
	req := httptest.NewRequest(http.MethodPost, "/organizations/"+orgID.String()+"/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_AddUser_InvalidRole(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	userID := uuid.New()

	body, _ := json.Marshal(map[string]string{
		"user_id": userID.String(),
		"role":    "invalid_role",
	})
	req := httptest.NewRequest(http.MethodPost, "/organizations/"+orgID.String()+"/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrganizationHandler_ListUsers_Success(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	users := []*model.OrganizationUser{
		{ID: uuid.New(), OrganizationID: orgID, UserID: uuid.New(), Role: model.OrgRoleAdmin},
		{ID: uuid.New(), OrganizationID: orgID, UserID: uuid.New(), Role: model.OrgRoleMember},
	}

	mockSvc.On("ListUsers", mock.Anything, orgID, mock.AnythingOfType("*model.ListOrgUsersRequest")).Return(users, int64(2), nil)

	req := httptest.NewRequest(http.MethodGet, "/organizations/"+orgID.String()+"/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_ListUsers_WithRoleFilter(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	users := []*model.OrganizationUser{
		{ID: uuid.New(), OrganizationID: orgID, UserID: uuid.New(), Role: model.OrgRoleAdmin},
	}

	mockSvc.On("ListUsers", mock.Anything, orgID, mock.MatchedBy(func(req *model.ListOrgUsersRequest) bool {
		return req.Role == model.OrgRoleAdmin
	})).Return(users, int64(1), nil)

	req := httptest.NewRequest(http.MethodGet, "/organizations/"+orgID.String()+"/users?role=org_admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_UpdateUserRole_Success(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	userID := uuid.New()
	orgUser := &model.OrganizationUser{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		Role:           model.OrgRoleAdmin,
	}

	mockSvc.On("UpdateUserRole", mock.Anything, orgID, userID, mock.AnythingOfType("*model.UpdateOrgUserRoleRequest")).Return(orgUser, nil)

	body, _ := json.Marshal(model.UpdateOrgUserRoleRequest{Role: model.OrgRoleAdmin})
	req := httptest.NewRequest(http.MethodPut, "/organizations/"+orgID.String()+"/users/"+userID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_UpdateUserRole_InvalidRole(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	userID := uuid.New()

	body, _ := json.Marshal(map[string]string{"role": "superadmin"})
	req := httptest.NewRequest(http.MethodPut, "/organizations/"+orgID.String()+"/users/"+userID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrganizationHandler_RemoveUser_Success(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	userID := uuid.New()

	mockSvc.On("RemoveUser", mock.Anything, orgID, userID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/"+orgID.String()+"/users/"+userID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_RemoveUser_CannotRemoveOwner(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	userID := uuid.New()

	mockSvc.On("RemoveUser", mock.Anything, orgID, userID).Return(service.ErrCannotRemoveOwner)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/"+orgID.String()+"/users/"+userID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_RemoveUser_NotInOrg(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	userID := uuid.New()

	mockSvc.On("RemoveUser", mock.Anything, orgID, userID).Return(service.ErrUserNotInOrg)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/"+orgID.String()+"/users/"+userID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_Update_InvalidBody(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	body := []byte(`invalid json`)
	req := httptest.NewRequest(http.MethodPut, "/organizations/"+orgID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrganizationHandler_List_InvalidQuery(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/organizations?page=-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrganizationHandler_List_ServiceError(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	mockSvc.On("List", mock.Anything, mock.Anything).Return(nil, int64(0), service.ErrOrganizationNotFound)

	req := httptest.NewRequest(http.MethodGet, "/organizations?page=1&page_size=20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_AddUser_InvalidOrgID(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	userID := uuid.New()
	body, _ := json.Marshal(model.AddOrgUserRequest{
		UserID: userID.String(),
		Role:   model.OrgRoleMember,
	})
	req := httptest.NewRequest(http.MethodPost, "/organizations/invalid-uuid/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrganizationHandler_AddUser_InvalidBody(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	body := []byte(`{"user_id": "invalid", "role": ""}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations/"+orgID.String()+"/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrganizationHandler_AddUser_OrgNotFound(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	userID := uuid.New()

	mockSvc.On("AddUser", mock.Anything, orgID, mock.Anything).Return(nil, service.ErrOrganizationNotFound)

	body, _ := json.Marshal(model.AddOrgUserRequest{
		UserID: userID.String(),
		Role:   model.OrgRoleMember,
	})
	req := httptest.NewRequest(http.MethodPost, "/organizations/"+orgID.String()+"/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_AddUser_UserNotFound(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	userID := uuid.New()

	mockSvc.On("AddUser", mock.Anything, orgID, mock.Anything).Return(nil, service.ErrUserNotFound)

	body, _ := json.Marshal(model.AddOrgUserRequest{
		UserID: userID.String(),
		Role:   model.OrgRoleMember,
	})
	req := httptest.NewRequest(http.MethodPost, "/organizations/"+orgID.String()+"/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_AddUser_InternalError(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	userID := uuid.New()

	mockSvc.On("AddUser", mock.Anything, orgID, mock.Anything).Return(nil, errors.New("internal error"))

	body, _ := json.Marshal(model.AddOrgUserRequest{
		UserID: userID.String(),
		Role:   model.OrgRoleMember,
	})
	req := httptest.NewRequest(http.MethodPost, "/organizations/"+orgID.String()+"/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_UpdateUserRole_InvalidOrgID(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	userID := uuid.New()
	body, _ := json.Marshal(model.UpdateOrgUserRoleRequest{Role: model.OrgRoleAdmin})
	req := httptest.NewRequest(http.MethodPut, "/organizations/invalid-uuid/users/"+userID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrganizationHandler_UpdateUserRole_InvalidUserID(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	body, _ := json.Marshal(model.UpdateOrgUserRoleRequest{Role: model.OrgRoleAdmin})
	req := httptest.NewRequest(http.MethodPut, "/organizations/"+orgID.String()+"/users/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrganizationHandler_UpdateUserRole_InvalidBody(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	userID := uuid.New()
	body := []byte(`{"role": ""}`)
	req := httptest.NewRequest(http.MethodPut, "/organizations/"+orgID.String()+"/users/"+userID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrganizationHandler_UpdateUserRole_OrgNotFound(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	userID := uuid.New()

	mockSvc.On("UpdateUserRole", mock.Anything, orgID, userID, mock.Anything).Return(nil, service.ErrOrganizationNotFound)

	body, _ := json.Marshal(model.UpdateOrgUserRoleRequest{Role: model.OrgRoleAdmin})
	req := httptest.NewRequest(http.MethodPut, "/organizations/"+orgID.String()+"/users/"+userID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_UpdateUserRole_UserNotInOrg(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	userID := uuid.New()

	mockSvc.On("UpdateUserRole", mock.Anything, orgID, userID, mock.Anything).Return(nil, service.ErrUserNotInOrg)

	body, _ := json.Marshal(model.UpdateOrgUserRoleRequest{Role: model.OrgRoleAdmin})
	req := httptest.NewRequest(http.MethodPut, "/organizations/"+orgID.String()+"/users/"+userID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_RemoveUser_InvalidOrgID(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/organizations/invalid-uuid/users/"+userID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrganizationHandler_RemoveUser_InvalidUserID(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/organizations/"+orgID.String()+"/users/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrganizationHandler_RemoveUser_OrgNotFound(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	userID := uuid.New()

	mockSvc.On("RemoveUser", mock.Anything, orgID, userID).Return(service.ErrOrganizationNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/"+orgID.String()+"/users/"+userID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_RemoveUser_InternalError(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	userID := uuid.New()

	mockSvc.On("RemoveUser", mock.Anything, orgID, userID).Return(errors.New("internal error"))

	req := httptest.NewRequest(http.MethodDelete, "/organizations/"+orgID.String()+"/users/"+userID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_ListUsers_InvalidOrgID(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/organizations/invalid-uuid/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrganizationHandler_ListUsers_InvalidQuery(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/organizations/"+orgID.String()+"/users?role=invalid_role", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrganizationHandler_ListUsers_OrgNotFound(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	mockSvc.On("ListUsers", mock.Anything, orgID, mock.Anything).Return(nil, int64(0), service.ErrOrganizationNotFound)

	req := httptest.NewRequest(http.MethodGet, "/organizations/"+orgID.String()+"/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_ListUsers_InternalError(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	mockSvc.On("ListUsers", mock.Anything, orgID, mock.Anything).Return(nil, int64(0), errors.New("internal error"))

	req := httptest.NewRequest(http.MethodGet, "/organizations/"+orgID.String()+"/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_Create_OwnerNotFound(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	ownerID := uuid.New()
	mockSvc.On("Create", mock.Anything, mock.Anything).Return(nil, service.ErrOwnerNotFound)

	body, _ := json.Marshal(model.CreateOrganizationRequest{
		Name:    "Test Org",
		OwnerID: ownerID.String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_Create_UserNotFound(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	ownerID := uuid.New()
	mockSvc.On("Create", mock.Anything, mock.Anything).Return(nil, service.ErrUserNotFound)

	body, _ := json.Marshal(model.CreateOrganizationRequest{
		Name:    "Test Org",
		OwnerID: ownerID.String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_Create_InternalError(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	ownerID := uuid.New()
	mockSvc.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("internal error"))

	body, _ := json.Marshal(model.CreateOrganizationRequest{
		Name:    "Test Org",
		OwnerID: ownerID.String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_Get_InternalError(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	mockSvc.On("GetByID", mock.Anything, orgID).Return(nil, errors.New("internal error"))

	req := httptest.NewRequest(http.MethodGet, "/organizations/"+orgID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_Update_InternalError(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	mockSvc.On("Update", mock.Anything, orgID, mock.Anything).Return(nil, errors.New("internal error"))

	body, _ := json.Marshal(model.UpdateOrganizationRequest{Name: "Updated Org"})
	req := httptest.NewRequest(http.MethodPut, "/organizations/"+orgID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestOrganizationHandler_Delete_InternalError(t *testing.T) {
	mockSvc := new(MockOrganizationService)
	h := NewOrganizationHandler(mockSvc)
	router := setupOrgRouter(h)

	orgID := uuid.New()
	mockSvc.On("Delete", mock.Anything, orgID).Return(errors.New("internal error"))

	req := httptest.NewRequest(http.MethodDelete, "/organizations/"+orgID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}
