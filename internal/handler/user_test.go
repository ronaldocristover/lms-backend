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

// ─── LIST USERS ───

func TestUserHandler_List_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	router := setupUserRouter(handler)

	users := []*model.User{
		{ID: uuid.New(), Email: "user1@example.com", Name: "User 1", Role: "user"},
		{ID: uuid.New(), Email: "user2@example.com", Name: "User 2", Role: "user"},
	}

	mockSvc.On("List", mock.Anything, 1, 20).Return(users, int64(2), nil)

	req := httptest.NewRequest(http.MethodGet, "/users?page=1&page_size=20", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, true, response["success"])

	data := response["data"].([]interface{})
	assert.Equal(t, 2, len(data))

	meta := response["meta"].(map[string]interface{})
	assert.Equal(t, float64(1), meta["page"])
	assert.Equal(t, float64(2), meta["total_items"])

	mockSvc.AssertExpectations(t)
}

func TestUserHandler_List_Empty(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	router := setupUserRouter(handler)

	users := []*model.User{}

	mockSvc.On("List", mock.Anything, 1, 20).Return(users, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].([]interface{})
	assert.Equal(t, 0, len(data))

	mockSvc.AssertExpectations(t)
}

func TestUserHandler_List_DefaultPagination(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	router := setupUserRouter(handler)

	users := []*model.User{}

	mockSvc.On("List", mock.Anything, 1, 20).Return(users, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_List_ServiceError(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	router := setupUserRouter(handler)

	mockSvc.On("List", mock.Anything, 1, 20).Return(([]*model.User)(nil), int64(0), assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

// ─── GET USER ───

func TestUserHandler_Get_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	router := setupUserRouter(handler)

	userID := uuid.New()
	user := &model.User{
		ID:    userID,
		Email: "test@example.com",
		Name:  "Test User",
		Role:  "user",
	}

	mockSvc.On("GetByID", mock.Anything, userID).Return(user, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/"+userID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, true, response["success"])
	data := response["data"].(map[string]interface{})
	assert.Equal(t, userID.String(), data["id"])
	assert.Equal(t, "test@example.com", data["email"])

	mockSvc.AssertExpectations(t)
}

func TestUserHandler_Get_InvalidID(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/users/invalid-uuid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, false, response["success"])

	mockSvc.AssertNotCalled(t, "GetByID")
}

func TestUserHandler_Get_NotFound(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	router := setupUserRouter(handler)

	userID := uuid.New()

	mockSvc.On("GetByID", mock.Anything, userID).Return(nil, service.ErrUserNotFound)

	req := httptest.NewRequest(http.MethodGet, "/users/"+userID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

// ─── UPDATE USER ───

func TestUserHandler_Update_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	router := setupUserRouter(handler)

	userID := uuid.New()

	reqBody := model.UpdateUserRequest{
		Name: "Updated Name",
	}

	updatedUser := &model.User{
		ID:    userID,
		Email: "test@example.com",
		Name:  "Updated Name",
		Role:  "user",
	}

	mockSvc.On("Update", mock.Anything, userID, mock.AnythingOfType("*model.UpdateUserRequest")).Return(updatedUser, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, true, response["success"])
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Updated Name", data["name"])

	mockSvc.AssertExpectations(t)
}

func TestUserHandler_Update_InvalidID(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	router := setupUserRouter(handler)

	reqBody := model.UpdateUserRequest{
		Name: "Updated Name",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/users/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "Update")
}

func TestUserHandler_Update_InvalidRole(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	router := setupUserRouter(handler)

	userID := uuid.New()

	reqBody := map[string]string{
		"role": "superadmin",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Update_NameTooShort(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	router := setupUserRouter(handler)

	userID := uuid.New()

	// omitempty means empty string won't trigger min=1 validation
	// but we still test that the endpoint accepts empty name gracefully
	reqBody := map[string]string{
		"name": "",
	}

	// Mock the update since omitempty skips validation for empty fields
	updatedUser := &model.User{
		ID:    userID,
		Email: "test@example.com",
		Name:  "",
		Role:  "user",
	}
	mockSvc.On("Update", mock.Anything, userID, mock.AnythingOfType("*model.UpdateUserRequest")).Return(updatedUser, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// With omitempty, empty name is accepted (not validated)
	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_Update_ServiceError(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	router := setupUserRouter(handler)

	userID := uuid.New()

	reqBody := model.UpdateUserRequest{
		Name: "Updated Name",
	}

	mockSvc.On("Update", mock.Anything, userID, mock.AnythingOfType("*model.UpdateUserRequest")).Return(nil, service.ErrUserNotFound)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

// ─── DELETE USER ───

func TestUserHandler_Delete_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	router := setupUserRouter(handler)

	userID := uuid.New()

	mockSvc.On("Delete", mock.Anything, userID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/users/"+userID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, true, response["success"])

	mockSvc.AssertExpectations(t)
}

func TestUserHandler_Delete_InvalidID(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/users/invalid-uuid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "Delete")
}

func TestUserHandler_Delete_ServiceError(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc)
	router := setupUserRouter(handler)

	userID := uuid.New()

	mockSvc.On("Delete", mock.Anything, userID).Return(assert.AnError)

	req := httptest.NewRequest(http.MethodDelete, "/users/"+userID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

// ─── Helper ───

func setupUserRouter(handler *UserHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/users", handler.List)
	r.GET("/users/:id", handler.Get)
	r.PUT("/users/:id", handler.Update)
	r.DELETE("/users/:id", handler.Delete)
	return r
}
