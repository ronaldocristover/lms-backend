package handler

import (
	"bytes"
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

func TestUserHandler_List_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewUserHandler(mockSvc)
	router := setupUserRouter(h)

	users := []*model.User{
		{ID: uuid.New(), Email: "user1@example.com", Name: "User 1", Role: model.UserRoleUser, Status: model.UserStatusActive},
		{ID: uuid.New(), Email: "user2@example.com", Name: "User 2", Role: model.UserRoleUser, Status: model.UserStatusActive},
	}

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListUsersRequest) bool {
		return req.Page == 1 && req.PageSize == 20
	})).Return(users, int64(2), nil)

	req := httptest.NewRequest(http.MethodGet, "/users?page=1&page_size=20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_List_WithFilter(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewUserHandler(mockSvc)
	router := setupUserRouter(h)

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListUsersRequest) bool {
		return req.Role == model.UserRoleTutor && req.Status == model.UserStatusActive
	})).Return([]*model.User{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/users?role=tutor&status=active", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_List_WithSearch(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewUserHandler(mockSvc)
	router := setupUserRouter(h)

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListUsersRequest) bool {
		return req.Search == "john"
	})).Return([]*model.User{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/users?search=john", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_List_Empty(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewUserHandler(mockSvc)
	router := setupUserRouter(h)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListUsersRequest")).Return([]*model.User{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_List_ServiceError(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewUserHandler(mockSvc)
	router := setupUserRouter(h)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListUsersRequest")).Return(([]*model.User)(nil), int64(0), assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_Get_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewUserHandler(mockSvc)
	router := setupUserRouter(h)

	userID := uuid.New()
	user := &model.User{ID: userID, Email: "test@example.com", Name: "Test User", Role: model.UserRoleUser, Status: model.UserStatusActive, Avatar: "https://example.com/avatar.png"}

	mockSvc.On("GetByID", mock.Anything, userID).Return(user, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/"+userID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	assert.Equal(t, userID.String(), data["id"])
	assert.Equal(t, "https://example.com/avatar.png", data["avatar"])
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_Get_InvalidID(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewUserHandler(mockSvc)
	router := setupUserRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/users/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Get_NotFound(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewUserHandler(mockSvc)
	router := setupUserRouter(h)

	userID := uuid.New()
	mockSvc.On("GetByID", mock.Anything, userID).Return(nil, service.ErrUserNotFound)

	req := httptest.NewRequest(http.MethodGet, "/users/"+userID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_Update_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewUserHandler(mockSvc)
	router := setupUserRouter(h)

	userID := uuid.New()
	updatedUser := &model.User{ID: userID, Email: "test@example.com", Name: "Updated Name", Role: model.UserRoleTutor, Status: model.UserStatusActive, Avatar: "https://example.com/new-avatar.png"}

	mockSvc.On("Update", mock.Anything, userID, mock.AnythingOfType("*model.UpdateUserRequest")).Return(updatedUser, nil)

	body, _ := json.Marshal(model.UpdateUserRequest{Name: "Updated Name", Role: model.UserRoleTutor, Avatar: "https://example.com/new-avatar.png"})
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_Update_Status(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewUserHandler(mockSvc)
	router := setupUserRouter(h)

	userID := uuid.New()
	updatedUser := &model.User{ID: userID, Email: "test@example.com", Status: model.UserStatusSuspended}

	mockSvc.On("Update", mock.Anything, userID, mock.AnythingOfType("*model.UpdateUserRequest")).Return(updatedUser, nil)

	body, _ := json.Marshal(model.UpdateUserRequest{Status: model.UserStatusSuspended})
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_Update_InvalidID(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewUserHandler(mockSvc)
	router := setupUserRouter(h)

	body, _ := json.Marshal(model.UpdateUserRequest{Name: "Test"})
	req := httptest.NewRequest(http.MethodPut, "/users/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Update_InvalidRole(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewUserHandler(mockSvc)
	router := setupUserRouter(h)

	userID := uuid.New()
	body, _ := json.Marshal(map[string]string{"role": "superadmin"})
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Update_InvalidStatus(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewUserHandler(mockSvc)
	router := setupUserRouter(h)

	userID := uuid.New()
	body, _ := json.Marshal(map[string]string{"status": "banned"})
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Update_ServiceError(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewUserHandler(mockSvc)
	router := setupUserRouter(h)

	userID := uuid.New()
	mockSvc.On("Update", mock.Anything, userID, mock.AnythingOfType("*model.UpdateUserRequest")).Return(nil, service.ErrUserNotFound)

	body, _ := json.Marshal(model.UpdateUserRequest{Name: "Test"})
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_Delete_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewUserHandler(mockSvc)
	router := setupUserRouter(h)

	userID := uuid.New()
	mockSvc.On("Delete", mock.Anything, userID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/users/"+userID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_Delete_InvalidID(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewUserHandler(mockSvc)
	router := setupUserRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/users/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Delete_ServiceError(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewUserHandler(mockSvc)
	router := setupUserRouter(h)

	userID := uuid.New()
	mockSvc.On("Delete", mock.Anything, userID).Return(assert.AnError)

	req := httptest.NewRequest(http.MethodDelete, "/users/"+userID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func setupUserRouter(handler *UserHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/users", handler.List)
	r.GET("/users/:id", handler.Get)
	r.PUT("/users/:id", handler.Update)
	r.DELETE("/users/:id", handler.Delete)
	return r
}
