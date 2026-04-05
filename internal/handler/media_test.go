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

type MockMediaService struct {
	mock.Mock
}

func (m *MockMediaService) Create(ctx context.Context, req *model.CreateMediaRequest) (*model.Media, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Media), args.Error(1)
}

func (m *MockMediaService) GetByID(ctx context.Context, id uuid.UUID) (*model.Media, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Media), args.Error(1)
}

func (m *MockMediaService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateMediaRequest) (*model.Media, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Media), args.Error(1)
}

func (m *MockMediaService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMediaService) List(ctx context.Context, req *model.ListMediaRequest) ([]*model.Media, int64, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Media), args.Get(1).(int64), args.Error(2)
}

func TestMediaHandler_Create_Success(t *testing.T) {
	mockSvc := new(MockMediaService)
	h := NewMediaHandler(mockSvc)
	router := setupMediaRouter(h)

	mediaID := uuid.New()
	created := &model.Media{ID: mediaID, Type: model.MediaTypeVideo, URL: "https://example.com/video.mp4"}

	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateMediaRequest")).Return(created, nil)

	body, _ := json.Marshal(model.CreateMediaRequest{Type: model.MediaTypeVideo, URL: "https://example.com/video.mp4"})
	req := httptest.NewRequest(http.MethodPost, "/media", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestMediaHandler_Create_InvalidBody(t *testing.T) {
	mockSvc := new(MockMediaService)
	h := NewMediaHandler(mockSvc)
	router := setupMediaRouter(h)

	body, _ := json.Marshal(map[string]string{"type": "invalid"})
	req := httptest.NewRequest(http.MethodPost, "/media", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaHandler_Get_Success(t *testing.T) {
	mockSvc := new(MockMediaService)
	h := NewMediaHandler(mockSvc)
	router := setupMediaRouter(h)

	mediaID := uuid.New()
	media := &model.Media{ID: mediaID, Type: model.MediaTypeVideo, URL: "https://example.com/video.mp4"}

	mockSvc.On("GetByID", mock.Anything, mediaID).Return(media, nil)

	req := httptest.NewRequest(http.MethodGet, "/media/"+mediaID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestMediaHandler_Get_InvalidID(t *testing.T) {
	mockSvc := new(MockMediaService)
	h := NewMediaHandler(mockSvc)
	router := setupMediaRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/media/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaHandler_Get_NotFound(t *testing.T) {
	mockSvc := new(MockMediaService)
	h := NewMediaHandler(mockSvc)
	router := setupMediaRouter(h)

	mediaID := uuid.New()
	mockSvc.On("GetByID", mock.Anything, mediaID).Return(nil, service.ErrMediaNotFound)

	req := httptest.NewRequest(http.MethodGet, "/media/"+mediaID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestMediaHandler_Update_Success(t *testing.T) {
	mockSvc := new(MockMediaService)
	h := NewMediaHandler(mockSvc)
	router := setupMediaRouter(h)

	mediaID := uuid.New()
	updated := &model.Media{ID: mediaID, Type: model.MediaTypeAudio, URL: "https://example.com/audio.mp3"}

	mockSvc.On("Update", mock.Anything, mediaID, mock.AnythingOfType("*model.UpdateMediaRequest")).Return(updated, nil)

	body, _ := json.Marshal(model.UpdateMediaRequest{URL: "https://example.com/audio.mp3"})
	req := httptest.NewRequest(http.MethodPut, "/media/"+mediaID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestMediaHandler_Update_InvalidID(t *testing.T) {
	mockSvc := new(MockMediaService)
	h := NewMediaHandler(mockSvc)
	router := setupMediaRouter(h)

	body, _ := json.Marshal(model.UpdateMediaRequest{URL: "https://example.com/new.mp4"})
	req := httptest.NewRequest(http.MethodPut, "/media/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaHandler_Update_NotFound(t *testing.T) {
	mockSvc := new(MockMediaService)
	h := NewMediaHandler(mockSvc)
	router := setupMediaRouter(h)

	mediaID := uuid.New()
	mockSvc.On("Update", mock.Anything, mediaID, mock.AnythingOfType("*model.UpdateMediaRequest")).Return(nil, service.ErrMediaNotFound)

	body, _ := json.Marshal(model.UpdateMediaRequest{URL: "https://example.com/new.mp4"})
	req := httptest.NewRequest(http.MethodPut, "/media/"+mediaID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestMediaHandler_Delete_Success(t *testing.T) {
	mockSvc := new(MockMediaService)
	h := NewMediaHandler(mockSvc)
	router := setupMediaRouter(h)

	mediaID := uuid.New()
	mockSvc.On("Delete", mock.Anything, mediaID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/media/"+mediaID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestMediaHandler_Delete_InvalidID(t *testing.T) {
	mockSvc := new(MockMediaService)
	h := NewMediaHandler(mockSvc)
	router := setupMediaRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/media/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaHandler_Delete_NotFound(t *testing.T) {
	mockSvc := new(MockMediaService)
	h := NewMediaHandler(mockSvc)
	router := setupMediaRouter(h)

	mediaID := uuid.New()
	mockSvc.On("Delete", mock.Anything, mediaID).Return(service.ErrMediaNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/media/"+mediaID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestMediaHandler_List_Success(t *testing.T) {
	mockSvc := new(MockMediaService)
	h := NewMediaHandler(mockSvc)
	router := setupMediaRouter(h)

	medias := []*model.Media{
		{ID: uuid.New(), Type: model.MediaTypeVideo, URL: "https://example.com/video1.mp4"},
		{ID: uuid.New(), Type: model.MediaTypeAudio, URL: "https://example.com/audio1.mp3"},
	}

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListMediaRequest) bool {
		return req.Page == 1 && req.PageSize == 20
	})).Return(medias, int64(2), nil)

	req := httptest.NewRequest(http.MethodGet, "/media?page=1&page_size=20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestMediaHandler_List_WithType(t *testing.T) {
	mockSvc := new(MockMediaService)
	h := NewMediaHandler(mockSvc)
	router := setupMediaRouter(h)

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListMediaRequest) bool {
		return req.Type == model.MediaTypeVideo
	})).Return([]*model.Media{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/media?type=video", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestMediaHandler_List_Empty(t *testing.T) {
	mockSvc := new(MockMediaService)
	h := NewMediaHandler(mockSvc)
	router := setupMediaRouter(h)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListMediaRequest")).Return([]*model.Media{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/media", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestMediaHandler_List_ServiceError(t *testing.T) {
	mockSvc := new(MockMediaService)
	h := NewMediaHandler(mockSvc)
	router := setupMediaRouter(h)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListMediaRequest")).Return(([]*model.Media)(nil), int64(0), assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/media", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func setupMediaRouter(handler *MediaHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/media", handler.Create)
	r.GET("/media", handler.List)
	r.GET("/media/:id", handler.Get)
	r.PUT("/media/:id", handler.Update)
	r.DELETE("/media/:id", handler.Delete)
	return r
}
