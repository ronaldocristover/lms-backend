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

type MockSubtitleService struct {
	mock.Mock
}

func (m *MockSubtitleService) Create(ctx context.Context, req *model.CreateSubtitleRequest) (*model.Subtitle, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subtitle), args.Error(1)
}

func (m *MockSubtitleService) GetByID(ctx context.Context, id uuid.UUID) (*model.Subtitle, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subtitle), args.Error(1)
}

func (m *MockSubtitleService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateSubtitleRequest) (*model.Subtitle, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subtitle), args.Error(1)
}

func (m *MockSubtitleService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSubtitleService) List(ctx context.Context, req *model.ListSubtitlesRequest) ([]*model.Subtitle, int64, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Subtitle), args.Get(1).(int64), args.Error(2)
}

func setupSubtitleRouter(handler *SubtitleHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/subtitles", handler.Create)
	r.GET("/subtitles", handler.List)
	r.GET("/subtitles/:id", handler.Get)
	r.PUT("/subtitles/:id", handler.Update)
	r.DELETE("/subtitles/:id", handler.Delete)
	return r
}

func TestSubtitleHandler_Create_Success(t *testing.T) {
	mockSvc := new(MockSubtitleService)
	h := NewSubtitleHandler(mockSvc)
	router := setupSubtitleRouter(h)

	subtitleID := uuid.New()
	mediaID := uuid.New()
	langID := uuid.New()
	created := &model.Subtitle{ID: subtitleID, MediaID: mediaID, LanguageID: langID, Content: "Hello world"}

	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateSubtitleRequest")).Return(created, nil)

	body, _ := json.Marshal(model.CreateSubtitleRequest{MediaID: mediaID, LanguageID: langID, Content: "Hello world"})
	req := httptest.NewRequest(http.MethodPost, "/subtitles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSubtitleHandler_Create_Duplicate(t *testing.T) {
	mockSvc := new(MockSubtitleService)
	h := NewSubtitleHandler(mockSvc)
	router := setupSubtitleRouter(h)

	mediaID := uuid.New()
	langID := uuid.New()
	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateSubtitleRequest")).Return(nil, service.ErrSubtitleExists)

	body, _ := json.Marshal(model.CreateSubtitleRequest{MediaID: mediaID, LanguageID: langID, Content: "Hello world"})
	req := httptest.NewRequest(http.MethodPost, "/subtitles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSubtitleHandler_Create_InvalidBody(t *testing.T) {
	mockSvc := new(MockSubtitleService)
	h := NewSubtitleHandler(mockSvc)
	router := setupSubtitleRouter(h)

	body, _ := json.Marshal(map[string]string{"content": "no media or language"})
	req := httptest.NewRequest(http.MethodPost, "/subtitles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubtitleHandler_Get_Success(t *testing.T) {
	mockSvc := new(MockSubtitleService)
	h := NewSubtitleHandler(mockSvc)
	router := setupSubtitleRouter(h)

	subtitleID := uuid.New()
	subtitle := &model.Subtitle{ID: subtitleID, MediaID: uuid.New(), LanguageID: uuid.New(), Content: "Test"}

	mockSvc.On("GetByID", mock.Anything, subtitleID).Return(subtitle, nil)

	req := httptest.NewRequest(http.MethodGet, "/subtitles/"+subtitleID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSubtitleHandler_Get_InvalidID(t *testing.T) {
	mockSvc := new(MockSubtitleService)
	h := NewSubtitleHandler(mockSvc)
	router := setupSubtitleRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/subtitles/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubtitleHandler_Get_NotFound(t *testing.T) {
	mockSvc := new(MockSubtitleService)
	h := NewSubtitleHandler(mockSvc)
	router := setupSubtitleRouter(h)

	subtitleID := uuid.New()
	mockSvc.On("GetByID", mock.Anything, subtitleID).Return(nil, service.ErrSubtitleNotFound)

	req := httptest.NewRequest(http.MethodGet, "/subtitles/"+subtitleID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSubtitleHandler_Update_Success(t *testing.T) {
	mockSvc := new(MockSubtitleService)
	h := NewSubtitleHandler(mockSvc)
	router := setupSubtitleRouter(h)

	subtitleID := uuid.New()
	updated := &model.Subtitle{ID: subtitleID, MediaID: uuid.New(), LanguageID: uuid.New(), Content: "Updated content"}

	mockSvc.On("Update", mock.Anything, subtitleID, mock.AnythingOfType("*model.UpdateSubtitleRequest")).Return(updated, nil)

	body, _ := json.Marshal(model.UpdateSubtitleRequest{Content: "Updated content"})
	req := httptest.NewRequest(http.MethodPut, "/subtitles/"+subtitleID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSubtitleHandler_Update_InvalidID(t *testing.T) {
	mockSvc := new(MockSubtitleService)
	h := NewSubtitleHandler(mockSvc)
	router := setupSubtitleRouter(h)

	body, _ := json.Marshal(model.UpdateSubtitleRequest{Content: "Updated content"})
	req := httptest.NewRequest(http.MethodPut, "/subtitles/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubtitleHandler_Update_NotFound(t *testing.T) {
	mockSvc := new(MockSubtitleService)
	h := NewSubtitleHandler(mockSvc)
	router := setupSubtitleRouter(h)

	subtitleID := uuid.New()
	mockSvc.On("Update", mock.Anything, subtitleID, mock.AnythingOfType("*model.UpdateSubtitleRequest")).Return(nil, service.ErrSubtitleNotFound)

	body, _ := json.Marshal(model.UpdateSubtitleRequest{Content: "Updated content"})
	req := httptest.NewRequest(http.MethodPut, "/subtitles/"+subtitleID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSubtitleHandler_Delete_Success(t *testing.T) {
	mockSvc := new(MockSubtitleService)
	h := NewSubtitleHandler(mockSvc)
	router := setupSubtitleRouter(h)

	subtitleID := uuid.New()
	mockSvc.On("Delete", mock.Anything, subtitleID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/subtitles/"+subtitleID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSubtitleHandler_Delete_InvalidID(t *testing.T) {
	mockSvc := new(MockSubtitleService)
	h := NewSubtitleHandler(mockSvc)
	router := setupSubtitleRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/subtitles/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubtitleHandler_Delete_NotFound(t *testing.T) {
	mockSvc := new(MockSubtitleService)
	h := NewSubtitleHandler(mockSvc)
	router := setupSubtitleRouter(h)

	subtitleID := uuid.New()
	mockSvc.On("Delete", mock.Anything, subtitleID).Return(service.ErrSubtitleNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/subtitles/"+subtitleID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSubtitleHandler_List_Success(t *testing.T) {
	mockSvc := new(MockSubtitleService)
	h := NewSubtitleHandler(mockSvc)
	router := setupSubtitleRouter(h)

	subtitles := []*model.Subtitle{
		{ID: uuid.New(), MediaID: uuid.New(), LanguageID: uuid.New(), Content: "Sub 1"},
		{ID: uuid.New(), MediaID: uuid.New(), LanguageID: uuid.New(), Content: "Sub 2"},
	}

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListSubtitlesRequest) bool {
		return req.Page == 1 && req.PageSize == 20
	})).Return(subtitles, int64(2), nil)

	req := httptest.NewRequest(http.MethodGet, "/subtitles?page=1&page_size=20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSubtitleHandler_List_WithMediaFilter(t *testing.T) {
	mockSvc := new(MockSubtitleService)
	h := NewSubtitleHandler(mockSvc)
	router := setupSubtitleRouter(h)

	mediaID := uuid.New()
	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListSubtitlesRequest) bool {
		return req.MediaID == mediaID.String()
	})).Return([]*model.Subtitle{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/subtitles?media_id="+mediaID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSubtitleHandler_List_Empty(t *testing.T) {
	mockSvc := new(MockSubtitleService)
	h := NewSubtitleHandler(mockSvc)
	router := setupSubtitleRouter(h)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListSubtitlesRequest")).Return([]*model.Subtitle{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/subtitles", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSubtitleHandler_List_ServiceError(t *testing.T) {
	mockSvc := new(MockSubtitleService)
	h := NewSubtitleHandler(mockSvc)
	router := setupSubtitleRouter(h)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListSubtitlesRequest")).Return(([]*model.Subtitle)(nil), int64(0), assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/subtitles", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}
