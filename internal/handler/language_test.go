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

type MockLanguageService struct {
	mock.Mock
}

func (m *MockLanguageService) Create(ctx context.Context, req *model.CreateLanguageRequest) (*model.Language, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Language), args.Error(1)
}

func (m *MockLanguageService) GetByID(ctx context.Context, id uuid.UUID) (*model.Language, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Language), args.Error(1)
}

func (m *MockLanguageService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateLanguageRequest) (*model.Language, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Language), args.Error(1)
}

func (m *MockLanguageService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockLanguageService) List(ctx context.Context, req *model.ListLanguagesRequest) ([]*model.Language, int64, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Language), args.Get(1).(int64), args.Error(2)
}

func TestLanguageHandler_Create_Success(t *testing.T) {
	mockSvc := new(MockLanguageService)
	h := NewLanguageHandler(mockSvc)
	router := setupLanguageRouter(h)

	languageID := uuid.New()
	created := &model.Language{ID: languageID, Code: "en", Name: "English"}

	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateLanguageRequest")).Return(created, nil)

	body, _ := json.Marshal(model.CreateLanguageRequest{Code: "en", Name: "English"})
	req := httptest.NewRequest(http.MethodPost, "/languages", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestLanguageHandler_Create_Duplicate(t *testing.T) {
	mockSvc := new(MockLanguageService)
	h := NewLanguageHandler(mockSvc)
	router := setupLanguageRouter(h)

	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateLanguageRequest")).Return(nil, service.ErrLanguageExists)

	body, _ := json.Marshal(model.CreateLanguageRequest{Code: "en", Name: "English"})
	req := httptest.NewRequest(http.MethodPost, "/languages", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestLanguageHandler_Create_InvalidLanguage(t *testing.T) {
	mockSvc := new(MockLanguageService)
	h := NewLanguageHandler(mockSvc)
	router := setupLanguageRouter(h)

	body, _ := json.Marshal(map[string]string{"code": "en"})
	req := httptest.NewRequest(http.MethodPost, "/languages", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLanguageHandler_Get_Success(t *testing.T) {
	mockSvc := new(MockLanguageService)
	h := NewLanguageHandler(mockSvc)
	router := setupLanguageRouter(h)

	languageID := uuid.New()
	language := &model.Language{ID: languageID, Code: "en", Name: "English"}

	mockSvc.On("GetByID", mock.Anything, languageID).Return(language, nil)

	req := httptest.NewRequest(http.MethodGet, "/languages/"+languageID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestLanguageHandler_Get_InvalidID(t *testing.T) {
	mockSvc := new(MockLanguageService)
	h := NewLanguageHandler(mockSvc)
	router := setupLanguageRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/languages/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLanguageHandler_Get_NotFound(t *testing.T) {
	mockSvc := new(MockLanguageService)
	h := NewLanguageHandler(mockSvc)
	router := setupLanguageRouter(h)

	languageID := uuid.New()
	mockSvc.On("GetByID", mock.Anything, languageID).Return(nil, service.ErrLanguageNotFound)

	req := httptest.NewRequest(http.MethodGet, "/languages/"+languageID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestLanguageHandler_Update_Success(t *testing.T) {
	mockSvc := new(MockLanguageService)
	h := NewLanguageHandler(mockSvc)
	router := setupLanguageRouter(h)

	languageID := uuid.New()
	updated := &model.Language{ID: languageID, Code: "es", Name: "Spanish"}

	mockSvc.On("Update", mock.Anything, languageID, mock.AnythingOfType("*model.UpdateLanguageRequest")).Return(updated, nil)

	body, _ := json.Marshal(model.UpdateLanguageRequest{Code: "es", Name: "Spanish"})
	req := httptest.NewRequest(http.MethodPut, "/languages/"+languageID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestLanguageHandler_Update_InvalidID(t *testing.T) {
	mockSvc := new(MockLanguageService)
	h := NewLanguageHandler(mockSvc)
	router := setupLanguageRouter(h)

	body, _ := json.Marshal(model.UpdateLanguageRequest{Code: "es"})
	req := httptest.NewRequest(http.MethodPut, "/languages/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLanguageHandler_Update_NotFound(t *testing.T) {
	mockSvc := new(MockLanguageService)
	h := NewLanguageHandler(mockSvc)
	router := setupLanguageRouter(h)

	languageID := uuid.New()
	mockSvc.On("Update", mock.Anything, languageID, mock.AnythingOfType("*model.UpdateLanguageRequest")).Return(nil, service.ErrLanguageNotFound)

	body, _ := json.Marshal(model.UpdateLanguageRequest{Code: "es"})
	req := httptest.NewRequest(http.MethodPut, "/languages/"+languageID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestLanguageHandler_Delete_Success(t *testing.T) {
	mockSvc := new(MockLanguageService)
	h := NewLanguageHandler(mockSvc)
	router := setupLanguageRouter(h)

	languageID := uuid.New()
	mockSvc.On("Delete", mock.Anything, languageID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/languages/"+languageID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestLanguageHandler_Delete_InvalidID(t *testing.T) {
	mockSvc := new(MockLanguageService)
	h := NewLanguageHandler(mockSvc)
	router := setupLanguageRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/languages/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLanguageHandler_Delete_NotFound(t *testing.T) {
	mockSvc := new(MockLanguageService)
	h := NewLanguageHandler(mockSvc)
	router := setupLanguageRouter(h)

	languageID := uuid.New()
	mockSvc.On("Delete", mock.Anything, languageID).Return(service.ErrLanguageNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/languages/"+languageID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestLanguageHandler_List_Success(t *testing.T) {
	mockSvc := new(MockLanguageService)
	h := NewLanguageHandler(mockSvc)
	router := setupLanguageRouter(h)

	languages := []*model.Language{
		{ID: uuid.New(), Code: "en", Name: "English"},
		{ID: uuid.New(), Code: "es", Name: "Spanish"},
	}

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListLanguagesRequest) bool {
		return req.Page == 1 && req.PageSize == 20
	})).Return(languages, int64(2), nil)

	req := httptest.NewRequest(http.MethodGet, "/languages?page=1&page_size=20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestLanguageHandler_List_WithSearch(t *testing.T) {
	mockSvc := new(MockLanguageService)
	h := NewLanguageHandler(mockSvc)
	router := setupLanguageRouter(h)

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListLanguagesRequest) bool {
		return req.Search == "en"
	})).Return([]*model.Language{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/languages?search=en", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestLanguageHandler_List_Empty(t *testing.T) {
	mockSvc := new(MockLanguageService)
	h := NewLanguageHandler(mockSvc)
	router := setupLanguageRouter(h)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListLanguagesRequest")).Return([]*model.Language{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/languages", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestLanguageHandler_List_ServiceError(t *testing.T) {
	mockSvc := new(MockLanguageService)
	h := NewLanguageHandler(mockSvc)
	router := setupLanguageRouter(h)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListLanguagesRequest")).Return(([]*model.Language)(nil), int64(0), assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/languages", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func setupLanguageRouter(handler *LanguageHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/languages", handler.Create)
	r.GET("/languages", handler.List)
	r.GET("/languages/:id", handler.Get)
	r.PUT("/languages/:id", handler.Update)
	r.DELETE("/languages/:id", handler.Delete)
	return r
}
