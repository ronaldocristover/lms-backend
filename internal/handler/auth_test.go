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

// ─── REGISTER ───

func TestAuthHandler_Register_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)
	router := setupRouter(h)

	reqBody := model.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	resp := &model.LoginResponse{
		Token: "jwt-token-here",
		User: model.User{
			ID:    uuid.New(),
			Email: reqBody.Email,
			Name:  reqBody.Name,
			Role:  "user",
		},
	}

	mockSvc.On("Register", mock.Anything, mock.AnythingOfType("*model.RegisterRequest")).Return(resp, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, true, response["success"])
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Register_ReturnsUser(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)
	router := setupRouter(h)

	userID := uuid.New()
	resp := &model.LoginResponse{
		Token: "token",
		User: model.User{ID: userID, Email: "new@example.com", Name: "New User", Role: "user"},
	}

	mockSvc.On("Register", mock.Anything, mock.AnythingOfType("*model.RegisterRequest")).Return(resp, nil)

	body, _ := json.Marshal(model.RegisterRequest{Email: "new@example.com", Password: "password123", Name: "New User"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	userData := data["user"].(map[string]interface{})
	assert.Equal(t, userID.String(), userData["id"])
	assert.Equal(t, "new@example.com", userData["email"])
}

func TestAuthHandler_Register_InvalidEmail(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)
	router := setupRouter(h)

	body, _ := json.Marshal(map[string]string{"email": "invalid-email", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_MissingPassword(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)
	router := setupRouter(h)

	body, _ := json.Marshal(map[string]string{"email": "test@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_ShortPassword(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)
	router := setupRouter(h)

	body, _ := json.Marshal(map[string]string{"email": "test@example.com", "password": "short"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_UserExists(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)
	router := setupRouter(h)

	mockSvc.On("Register", mock.Anything, mock.AnythingOfType("*model.RegisterRequest")).Return(nil, service.ErrUserExists)

	body, _ := json.Marshal(model.RegisterRequest{Email: "existing@example.com", Password: "password123", Name: "Existing"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Register_EmptyBody(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)
	router := setupRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_InternalError(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)
	router := setupRouter(h)

	mockSvc.On("Register", mock.Anything, mock.AnythingOfType("*model.RegisterRequest")).Return(nil, context.DeadlineExceeded)

	body, _ := json.Marshal(model.RegisterRequest{Email: "test@example.com", Password: "password123", Name: "Test"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

// ─── LOGIN ───

func TestAuthHandler_Login_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)
	router := setupRouter(h)

	resp := &model.LoginResponse{
		Token: "jwt-token",
		User:  model.User{ID: uuid.New(), Email: "test@example.com", Name: "Test User", Role: "user"},
	}

	mockSvc.On("Login", mock.Anything, mock.AnythingOfType("*model.LoginRequest")).Return(resp, nil)

	body, _ := json.Marshal(model.LoginRequest{Email: "test@example.com", Password: "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, true, response["success"])
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)
	router := setupRouter(h)

	mockSvc.On("Login", mock.Anything, mock.AnythingOfType("*model.LoginRequest")).Return(nil, service.ErrInvalidCredentials)

	body, _ := json.Marshal(model.LoginRequest{Email: "test@example.com", Password: "wrong"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidEmail(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)
	router := setupRouter(h)

	body, _ := json.Marshal(map[string]string{"email": "not-an-email", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_MissingFields(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)
	router := setupRouter(h)

	body, _ := json.Marshal(map[string]string{"email": "test@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_EmptyBody(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)
	router := setupRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_InternalError(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)
	router := setupRouter(h)

	mockSvc.On("Login", mock.Anything, mock.AnythingOfType("*model.LoginRequest")).Return(nil, context.DeadlineExceeded)

	body, _ := json.Marshal(model.LoginRequest{Email: "test@example.com", Password: "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

// ─── ME (Current User) ───

func TestAuthHandler_Me_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)
	router := gin.New()

	userID := uuid.New()
	user := &model.User{ID: userID, Email: "test@example.com", Name: "Test User", Role: "user"}

	mockSvc.On("GetByID", mock.Anything, userID).Return(user, nil)

	router.Use(func(c *gin.Context) { c.Set("userID", userID); c.Next() })
	router.GET("/auth/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, true, response["success"])
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Me_UserNotAuthenticated(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)
	router := gin.New()
	router.GET("/auth/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Me_UserNotFound(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)
	router := gin.New()

	userID := uuid.New()
	mockSvc.On("GetByID", mock.Anything, userID).Return(nil, service.ErrUserNotFound)

	router.Use(func(c *gin.Context) { c.Set("userID", userID); c.Next() })
	router.GET("/auth/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}
