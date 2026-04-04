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
	"golang.org/x/crypto/bcrypt"
)

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, req *model.RegisterRequest) (*model.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.LoginResponse), args.Error(1)
}

func (m *MockUserService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.LoginResponse), args.Error(1)
}

func (m *MockUserService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateUserRequest) (*model.User, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) List(ctx context.Context, page, pageSize int) ([]*model.User, int64, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]*model.User), args.Get(1).(int64), args.Error(2)
}

func setupRouter(handler *AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/register", handler.Register)
	r.POST("/auth/login", handler.Login)
	return r
}

func TestAuthHandler_Register_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewAuthHandler(mockSvc)
	router := setupRouter(handler)

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

func TestAuthHandler_Register_ValidationError(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewAuthHandler(mockSvc)
	router := setupRouter(handler)

	// Invalid email
	reqBody := map[string]string{
		"email":    "invalid-email",
		"password": "password123",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_UserExists(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewAuthHandler(mockSvc)
	router := setupRouter(handler)

	reqBody := model.RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}

	mockSvc.On("Register", mock.Anything, mock.AnythingOfType("*model.RegisterRequest")).Return(nil, service.ErrUserExists)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewAuthHandler(mockSvc)
	router := setupRouter(handler)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	reqBody := model.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	resp := &model.LoginResponse{
		Token: "jwt-token-here",
		User: model.User{
			ID:           uuid.New(),
			Email:        reqBody.Email,
			PasswordHash: string(hashedPassword),
			Role:         "user",
		},
	}

	mockSvc.On("Login", mock.Anything, mock.AnythingOfType("*model.LoginRequest")).Return(resp, nil)

	body, _ := json.Marshal(reqBody)
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
	handler := NewAuthHandler(mockSvc)
	router := setupRouter(handler)

	reqBody := model.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	mockSvc.On("Login", mock.Anything, mock.AnythingOfType("*model.LoginRequest")).Return(nil, service.ErrInvalidCredentials)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Me_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	handler := NewAuthHandler(mockSvc)
	router := gin.New()

	userID := uuid.New()
	user := &model.User{
		ID:    userID,
		Email: "test@example.com",
		Name:  "Test User",
		Role:  "user",
	}

	mockSvc.On("GetByID", mock.Anything, userID).Return(user, nil)

	// Create a custom middleware that sets userID in context
	router.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	})
	router.GET("/auth/me", handler.Me)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
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
	handler := NewAuthHandler(mockSvc)
	router := gin.New()

	// No middleware that sets userID
	router.GET("/auth/me", handler.Me)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", HealthCheck)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "ok", response["status"])
}
