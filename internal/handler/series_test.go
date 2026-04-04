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

type MockSeriesService struct {
	mock.Mock
}

func (m *MockSeriesService) Create(ctx context.Context, req *model.CreateSeriesRequest) (*model.Series, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Series), args.Error(1)
}

func (m *MockSeriesService) GetByID(ctx context.Context, id uuid.UUID) (*model.Series, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Series), args.Error(1)
}

func (m *MockSeriesService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateSeriesRequest) (*model.Series, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Series), args.Error(1)
}

func (m *MockSeriesService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSeriesService) List(ctx context.Context, req *model.ListSeriesRequest) ([]*model.Series, int64, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Series), args.Get(1).(int64), args.Error(2)
}

func TestSeriesHandler_Create_Success(t *testing.T) {
	mockSvc := new(MockSeriesService)
	h := NewSeriesHandler(mockSvc)
	router := setupSeriesRouter(h)

	seriesID := uuid.New()
	categoryID := uuid.New()
	created := &model.Series{ID: seriesID, Title: "Go Programming", CategoryID: categoryID, IsPaid: false}

	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateSeriesRequest")).Return(created, nil)

	body, _ := json.Marshal(model.CreateSeriesRequest{Title: "Go Programming", CategoryID: categoryID, IsPaid: false})
	req := httptest.NewRequest(http.MethodPost, "/series", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSeriesHandler_Create_CategoryNotFound(t *testing.T) {
	mockSvc := new(MockSeriesService)
	h := NewSeriesHandler(mockSvc)
	router := setupSeriesRouter(h)

	categoryID := uuid.New()
	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateSeriesRequest")).Return(nil, service.ErrCategoryNotFound)

	body, _ := json.Marshal(model.CreateSeriesRequest{Title: "Go Programming", CategoryID: categoryID})
	req := httptest.NewRequest(http.MethodPost, "/series", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSeriesHandler_Create_InvalidInput(t *testing.T) {
	mockSvc := new(MockSeriesService)
	h := NewSeriesHandler(mockSvc)
	router := setupSeriesRouter(h)

	body, _ := json.Marshal(map[string]string{"invalid": "data"})
	req := httptest.NewRequest(http.MethodPost, "/series", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSeriesHandler_Get_Success(t *testing.T) {
	mockSvc := new(MockSeriesService)
	h := NewSeriesHandler(mockSvc)
	router := setupSeriesRouter(h)

	seriesID := uuid.New()
	series := &model.Series{ID: seriesID, Title: "Go Programming"}

	mockSvc.On("GetByID", mock.Anything, seriesID).Return(series, nil)

	req := httptest.NewRequest(http.MethodGet, "/series/"+seriesID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSeriesHandler_Get_InvalidID(t *testing.T) {
	mockSvc := new(MockSeriesService)
	h := NewSeriesHandler(mockSvc)
	router := setupSeriesRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/series/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSeriesHandler_Get_NotFound(t *testing.T) {
	mockSvc := new(MockSeriesService)
	h := NewSeriesHandler(mockSvc)
	router := setupSeriesRouter(h)

	seriesID := uuid.New()
	mockSvc.On("GetByID", mock.Anything, seriesID).Return(nil, service.ErrSeriesNotFound)

	req := httptest.NewRequest(http.MethodGet, "/series/"+seriesID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSeriesHandler_Update_Success(t *testing.T) {
	mockSvc := new(MockSeriesService)
	h := NewSeriesHandler(mockSvc)
	router := setupSeriesRouter(h)

	seriesID := uuid.New()
	categoryID := uuid.New()
	updated := &model.Series{ID: seriesID, Title: "Updated Series", CategoryID: categoryID, IsPaid: true}

	mockSvc.On("Update", mock.Anything, seriesID, mock.AnythingOfType("*model.UpdateSeriesRequest")).Return(updated, nil)

	body, _ := json.Marshal(model.UpdateSeriesRequest{Title: "Updated Series", CategoryID: categoryID, IsPaid: true})
	req := httptest.NewRequest(http.MethodPut, "/series/"+seriesID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSeriesHandler_Update_InvalidID(t *testing.T) {
	mockSvc := new(MockSeriesService)
	h := NewSeriesHandler(mockSvc)
	router := setupSeriesRouter(h)

	categoryID := uuid.New()
	body, _ := json.Marshal(model.UpdateSeriesRequest{Title: "Updated Series", CategoryID: categoryID})
	req := httptest.NewRequest(http.MethodPut, "/series/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSeriesHandler_Update_NotFound(t *testing.T) {
	mockSvc := new(MockSeriesService)
	h := NewSeriesHandler(mockSvc)
	router := setupSeriesRouter(h)

	seriesID := uuid.New()
	categoryID := uuid.New()
	mockSvc.On("Update", mock.Anything, seriesID, mock.AnythingOfType("*model.UpdateSeriesRequest")).Return(nil, service.ErrSeriesNotFound)

	body, _ := json.Marshal(model.UpdateSeriesRequest{Title: "Updated Series", CategoryID: categoryID})
	req := httptest.NewRequest(http.MethodPut, "/series/"+seriesID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSeriesHandler_Delete_Success(t *testing.T) {
	mockSvc := new(MockSeriesService)
	h := NewSeriesHandler(mockSvc)
	router := setupSeriesRouter(h)

	seriesID := uuid.New()
	mockSvc.On("Delete", mock.Anything, seriesID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/series/"+seriesID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSeriesHandler_Delete_InvalidID(t *testing.T) {
	mockSvc := new(MockSeriesService)
	h := NewSeriesHandler(mockSvc)
	router := setupSeriesRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/series/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSeriesHandler_Delete_NotFound(t *testing.T) {
	mockSvc := new(MockSeriesService)
	h := NewSeriesHandler(mockSvc)
	router := setupSeriesRouter(h)

	seriesID := uuid.New()
	mockSvc.On("Delete", mock.Anything, seriesID).Return(service.ErrSeriesNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/series/"+seriesID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSeriesHandler_List_Success(t *testing.T) {
	mockSvc := new(MockSeriesService)
	h := NewSeriesHandler(mockSvc)
	router := setupSeriesRouter(h)

	series := []*model.Series{
		{ID: uuid.New(), Title: "Go Programming"},
		{ID: uuid.New(), Title: "Python Basics"},
	}

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListSeriesRequest) bool {
		return req.Page == 1 && req.PageSize == 20
	})).Return(series, int64(2), nil)

	req := httptest.NewRequest(http.MethodGet, "/series?page=1&page_size=20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSeriesHandler_List_WithSearch(t *testing.T) {
	mockSvc := new(MockSeriesService)
	h := NewSeriesHandler(mockSvc)
	router := setupSeriesRouter(h)

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListSeriesRequest) bool {
		return req.Search == "go"
	})).Return([]*model.Series{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/series?search=go", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSeriesHandler_List_Empty(t *testing.T) {
	mockSvc := new(MockSeriesService)
	h := NewSeriesHandler(mockSvc)
	router := setupSeriesRouter(h)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListSeriesRequest")).Return([]*model.Series{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/series", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSeriesHandler_List_ServiceError(t *testing.T) {
	mockSvc := new(MockSeriesService)
	h := NewSeriesHandler(mockSvc)
	router := setupSeriesRouter(h)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListSeriesRequest")).Return(([]*model.Series)(nil), int64(0), assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/series", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func setupSeriesRouter(handler *SeriesHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/series", handler.Create)
	r.GET("/series", handler.List)
	r.GET("/series/:id", handler.Get)
	r.PUT("/series/:id", handler.Update)
	r.DELETE("/series/:id", handler.Delete)
	return r
}
