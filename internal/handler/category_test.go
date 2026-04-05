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

type MockCategoryService struct {
	mock.Mock
}

func (m *MockCategoryService) Create(ctx context.Context, req *model.CreateCategoryRequest) (*model.Category, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Category), args.Error(1)
}

func (m *MockCategoryService) GetByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Category), args.Error(1)
}

func (m *MockCategoryService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateCategoryRequest) (*model.Category, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Category), args.Error(1)
}

func (m *MockCategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCategoryService) List(ctx context.Context, req *model.ListCategoriesRequest) ([]*model.Category, int64, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Category), args.Get(1).(int64), args.Error(2)
}

func TestCategoryHandler_Create_Success(t *testing.T) {
	mockSvc := new(MockCategoryService)
	h := NewCategoryHandler(mockSvc)
	router := setupCategoryRouter(h)

	categoryID := uuid.New()
	created := &model.Category{ID: categoryID, Name: "Programming"}

	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateCategoryRequest")).Return(created, nil)

	body, _ := json.Marshal(model.CreateCategoryRequest{Name: "Programming"})
	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestCategoryHandler_Create_Duplicate(t *testing.T) {
	mockSvc := new(MockCategoryService)
	h := NewCategoryHandler(mockSvc)
	router := setupCategoryRouter(h)

	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateCategoryRequest")).Return(nil, service.ErrCategoryExists)

	body, _ := json.Marshal(model.CreateCategoryRequest{Name: "Programming"})
	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestCategoryHandler_Create_InvalidInput(t *testing.T) {
	mockSvc := new(MockCategoryService)
	h := NewCategoryHandler(mockSvc)
	router := setupCategoryRouter(h)

	body, _ := json.Marshal(map[string]string{"invalid": "data"})
	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCategoryHandler_Get_Success(t *testing.T) {
	mockSvc := new(MockCategoryService)
	h := NewCategoryHandler(mockSvc)
	router := setupCategoryRouter(h)

	categoryID := uuid.New()
	category := &model.Category{ID: categoryID, Name: "Programming"}

	mockSvc.On("GetByID", mock.Anything, categoryID).Return(category, nil)

	req := httptest.NewRequest(http.MethodGet, "/categories/"+categoryID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestCategoryHandler_Get_InvalidID(t *testing.T) {
	mockSvc := new(MockCategoryService)
	h := NewCategoryHandler(mockSvc)
	router := setupCategoryRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/categories/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCategoryHandler_Get_NotFound(t *testing.T) {
	mockSvc := new(MockCategoryService)
	h := NewCategoryHandler(mockSvc)
	router := setupCategoryRouter(h)

	categoryID := uuid.New()
	mockSvc.On("GetByID", mock.Anything, categoryID).Return(nil, service.ErrCategoryNotFound)

	req := httptest.NewRequest(http.MethodGet, "/categories/"+categoryID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestCategoryHandler_Update_Success(t *testing.T) {
	mockSvc := new(MockCategoryService)
	h := NewCategoryHandler(mockSvc)
	router := setupCategoryRouter(h)

	categoryID := uuid.New()
	updated := &model.Category{ID: categoryID, Name: "Updated Category"}

	mockSvc.On("Update", mock.Anything, categoryID, mock.AnythingOfType("*model.UpdateCategoryRequest")).Return(updated, nil)

	body, _ := json.Marshal(model.UpdateCategoryRequest{Name: "Updated Category"})
	req := httptest.NewRequest(http.MethodPut, "/categories/"+categoryID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestCategoryHandler_Update_InvalidID(t *testing.T) {
	mockSvc := new(MockCategoryService)
	h := NewCategoryHandler(mockSvc)
	router := setupCategoryRouter(h)

	body, _ := json.Marshal(model.UpdateCategoryRequest{Name: "Updated Category"})
	req := httptest.NewRequest(http.MethodPut, "/categories/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCategoryHandler_Update_NotFound(t *testing.T) {
	mockSvc := new(MockCategoryService)
	h := NewCategoryHandler(mockSvc)
	router := setupCategoryRouter(h)

	categoryID := uuid.New()
	mockSvc.On("Update", mock.Anything, categoryID, mock.AnythingOfType("*model.UpdateCategoryRequest")).Return(nil, service.ErrCategoryNotFound)

	body, _ := json.Marshal(model.UpdateCategoryRequest{Name: "Updated Category"})
	req := httptest.NewRequest(http.MethodPut, "/categories/"+categoryID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestCategoryHandler_Delete_Success(t *testing.T) {
	mockSvc := new(MockCategoryService)
	h := NewCategoryHandler(mockSvc)
	router := setupCategoryRouter(h)

	categoryID := uuid.New()
	mockSvc.On("Delete", mock.Anything, categoryID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/categories/"+categoryID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestCategoryHandler_Delete_InvalidID(t *testing.T) {
	mockSvc := new(MockCategoryService)
	h := NewCategoryHandler(mockSvc)
	router := setupCategoryRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/categories/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCategoryHandler_Delete_NotFound(t *testing.T) {
	mockSvc := new(MockCategoryService)
	h := NewCategoryHandler(mockSvc)
	router := setupCategoryRouter(h)

	categoryID := uuid.New()
	mockSvc.On("Delete", mock.Anything, categoryID).Return(service.ErrCategoryNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/categories/"+categoryID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestCategoryHandler_List_Success(t *testing.T) {
	mockSvc := new(MockCategoryService)
	h := NewCategoryHandler(mockSvc)
	router := setupCategoryRouter(h)

	categories := []*model.Category{
		{ID: uuid.New(), Name: "Programming"},
		{ID: uuid.New(), Name: "Design"},
	}

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListCategoriesRequest) bool {
		return req.Page == 1 && req.PageSize == 20
	})).Return(categories, int64(2), nil)

	req := httptest.NewRequest(http.MethodGet, "/categories?page=1&page_size=20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestCategoryHandler_List_WithSearch(t *testing.T) {
	mockSvc := new(MockCategoryService)
	h := NewCategoryHandler(mockSvc)
	router := setupCategoryRouter(h)

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListCategoriesRequest) bool {
		return req.Search == "programming"
	})).Return([]*model.Category{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/categories?search=programming", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestCategoryHandler_List_Empty(t *testing.T) {
	mockSvc := new(MockCategoryService)
	h := NewCategoryHandler(mockSvc)
	router := setupCategoryRouter(h)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListCategoriesRequest")).Return([]*model.Category{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestCategoryHandler_List_ServiceError(t *testing.T) {
	mockSvc := new(MockCategoryService)
	h := NewCategoryHandler(mockSvc)
	router := setupCategoryRouter(h)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListCategoriesRequest")).Return(([]*model.Category)(nil), int64(0), assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func setupCategoryRouter(handler *CategoryHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/categories", handler.Create)
	r.GET("/categories", handler.List)
	r.GET("/categories/:id", handler.Get)
	r.PUT("/categories/:id", handler.Update)
	r.DELETE("/categories/:id", handler.Delete)
	return r
}
