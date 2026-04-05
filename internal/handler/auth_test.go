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

	"github.com/ronaldocristover/lms-backend/internal/model"
	"github.com/ronaldocristover/lms-backend/internal/service"
)

func TestAuthHandler_Register_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
	router := setupRouter(h)

	roleID := uuid.New()
	reqBody := model.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		RoleID:   roleID,
	}

	resp := &model.LoginResponse{
		Token: "jwt-token-here",
		User: model.User{
			ID:     uuid.New(),
			Email:  reqBody.Email,
			Name:   reqBody.Name,
			RoleID: roleID,
			Role:   &model.Role{ID: roleID, Name: model.RoleStudent},
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
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
	router := setupRouter(h)

	userID := uuid.New()
	roleID := uuid.New()
	resp := &model.LoginResponse{
		Token: "token",
		User: model.User{
			ID:     userID,
			Email:  "new@example.com",
			Name:   "New User",
			RoleID: roleID,
			Role:   &model.Role{ID: roleID, Name: model.RoleStudent},
		},
	}

	mockSvc.On("Register", mock.Anything, mock.AnythingOfType("*model.RegisterRequest")).Return(resp, nil)

	body, _ := json.Marshal(model.RegisterRequest{Email: "new@example.com", Password: "password123", Name: "New User", RoleID: roleID})
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
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
	router := setupRouter(h)

	body, _ := json.Marshal(map[string]string{"email": "invalid-email", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_MissingPassword(t *testing.T) {
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
	router := setupRouter(h)

	body, _ := json.Marshal(map[string]string{"email": "test@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_ShortPassword(t *testing.T) {
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
	router := setupRouter(h)

	body, _ := json.Marshal(map[string]string{"email": "test@example.com", "password": "short"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_UserExists(t *testing.T) {
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
	router := setupRouter(h)

	mockSvc.On("Register", mock.Anything, mock.AnythingOfType("*model.RegisterRequest")).Return(nil, service.ErrUserExists)

	roleID := uuid.New()
	body, _ := json.Marshal(model.RegisterRequest{Email: "existing@example.com", Password: "password123", Name: "Existing", RoleID: roleID})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Register_EmptyBody(t *testing.T) {
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
	router := setupRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_InternalError(t *testing.T) {
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
	router := setupRouter(h)

	mockSvc.On("Register", mock.Anything, mock.AnythingOfType("*model.RegisterRequest")).Return(nil, context.DeadlineExceeded)

	roleID := uuid.New()
	body, _ := json.Marshal(model.RegisterRequest{Email: "test@example.com", Password: "password123", Name: "Test", RoleID: roleID})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
	router := setupRouter(h)

	roleID := uuid.New()
	resp := &model.LoginResponse{
		Token: "jwt-token",
		User: model.User{
			ID:     uuid.New(),
			Email:  "test@example.com",
			Name:   "Test User",
			RoleID: roleID,
			Role:   &model.Role{ID: roleID, Name: model.RoleStudent},
		},
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
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
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
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
	router := setupRouter(h)

	body, _ := json.Marshal(map[string]string{"email": "not-an-email", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_MissingFields(t *testing.T) {
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
	router := setupRouter(h)

	body, _ := json.Marshal(map[string]string{"email": "test@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_EmptyBody(t *testing.T) {
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
	router := setupRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_InternalError(t *testing.T) {
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
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

func TestAuthHandler_Me_Success(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)
	h := NewAuthHandler(mockAuthSvc, mockUserSvc)
	router := gin.New()

	userID := uuid.New()
	roleID := uuid.New()
	user := &model.User{
		ID:     userID,
		Email:  "test@example.com",
		Name:   "Test User",
		RoleID: roleID,
		Role:   &model.Role{ID: roleID, Name: model.RoleStudent},
	}

	mockUserSvc.On("GetByID", mock.Anything, userID).Return(user, nil)

	router.Use(func(c *gin.Context) { c.Set("userID", userID); c.Next() })
	router.GET("/auth/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, true, response["success"])
	mockAuthSvc.AssertExpectations(t)
}

func TestAuthHandler_Me_UserNotAuthenticated(t *testing.T) {
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
	router := gin.New()
	router.GET("/auth/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Me_UserNotFound(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockUserSvc := new(MockUserService)
	h := NewAuthHandler(mockAuthSvc, mockUserSvc)
	router := gin.New()

	userID := uuid.New()
	mockUserSvc.On("GetByID", mock.Anything, userID).Return(nil, service.ErrUserNotFound)

	router.Use(func(c *gin.Context) { c.Set("userID", userID); c.Next() })
	router.GET("/auth/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUserSvc.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
	router := setupRouter(h)

	userID := uuid.New()
	resp := &model.LoginResponse{
		Token:        "new-access-token",
		RefreshToken: "new-refresh-token",
		User:         model.User{ID: userID, Email: "test@example.com"},
	}

	mockSvc.On("RefreshToken", mock.Anything, "valid-refresh-token").Return(resp, nil)

	body, _ := json.Marshal(model.RefreshTokenRequest{RefreshToken: "valid-refresh-token"})
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, true, response["success"])
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "new-access-token", data["token"])
	assert.Equal(t, "new-refresh-token", data["refresh_token"])
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_MissingBody(t *testing.T) {
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
	router := setupRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_RefreshToken_InvalidToken(t *testing.T) {
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
	router := setupRouter(h)

	mockSvc.On("RefreshToken", mock.Anything, "bad-token").Return(nil, service.ErrInvalidToken)

	body, _ := json.Marshal(model.RefreshTokenRequest{RefreshToken: "bad-token"})
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_UserNotFound(t *testing.T) {
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
	router := setupRouter(h)

	mockSvc.On("RefreshToken", mock.Anything, "orphan-token").Return(nil, service.ErrUserNotFound)

	body, _ := json.Marshal(model.RefreshTokenRequest{RefreshToken: "orphan-token"})
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_InternalError(t *testing.T) {
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc, new(MockUserService))
	router := setupRouter(h)

	mockSvc.On("RefreshToken", mock.Anything, "some-token").Return(nil, assert.AnError)

	body, _ := json.Marshal(model.RefreshTokenRequest{RefreshToken: "some-token"})
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}
