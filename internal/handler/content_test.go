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

type MockContentService struct {
	mock.Mock
}

func (m *MockContentService) Create(ctx context.Context, req *model.CreateContentRequest) (*model.Content, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Content), args.Error(1)
}

func (m *MockContentService) GetByID(ctx context.Context, id uuid.UUID) (*model.Content, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Content), args.Error(1)
}

func (m *MockContentService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateContentRequest) (*model.Content, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Content), args.Error(1)
}

func (m *MockContentService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContentService) List(ctx context.Context, req *model.ListContentsRequest) ([]*model.Content, int64, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Content), args.Get(1).(int64), args.Error(2)
}

func TestContentHandler_Create_Success(t *testing.T) {
	mockSvc := new(MockContentService)
	h := NewContentHandler(mockSvc)
	router := setupContentRouter(h)

	contentID := uuid.New()
	sessionID := uuid.New()
	created := &model.Content{ID: contentID, SessionID: sessionID, Type: model.ContentTypeVideo, ContentText: "Sample content"}

	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateContentRequest")).Return(created, nil)

	body, _ := json.Marshal(model.CreateContentRequest{SessionID: sessionID, Type: model.ContentTypeVideo, ContentText: "Sample content"})
	req := httptest.NewRequest(http.MethodPost, "/contents", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestContentHandler_Create_SessionNotFound(t *testing.T) {
	mockSvc := new(MockContentService)
	h := NewContentHandler(mockSvc)
	router := setupContentRouter(h)

	sessionID := uuid.New()
	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateContentRequest")).Return(nil, service.ErrSessionNotFound)

	body, _ := json.Marshal(model.CreateContentRequest{SessionID: sessionID, Type: model.ContentTypeVideo})
	req := httptest.NewRequest(http.MethodPost, "/contents", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestContentHandler_Create_InvalidInput(t *testing.T) {
	mockSvc := new(MockContentService)
	h := NewContentHandler(mockSvc)
	router := setupContentRouter(h)

	body, _ := json.Marshal(map[string]string{"invalid": "data"})
	req := httptest.NewRequest(http.MethodPost, "/contents", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContentHandler_Get_Success(t *testing.T) {
	mockSvc := new(MockContentService)
	h := NewContentHandler(mockSvc)
	router := setupContentRouter(h)

	contentID := uuid.New()
	content := &model.Content{ID: contentID, Type: model.ContentTypeVideo}

	mockSvc.On("GetByID", mock.Anything, contentID).Return(content, nil)

	req := httptest.NewRequest(http.MethodGet, "/contents/"+contentID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestContentHandler_Get_InvalidID(t *testing.T) {
	mockSvc := new(MockContentService)
	h := NewContentHandler(mockSvc)
	router := setupContentRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/contents/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContentHandler_Get_NotFound(t *testing.T) {
	mockSvc := new(MockContentService)
	h := NewContentHandler(mockSvc)
	router := setupContentRouter(h)

	contentID := uuid.New()
	mockSvc.On("GetByID", mock.Anything, contentID).Return(nil, service.ErrContentNotFound)

	req := httptest.NewRequest(http.MethodGet, "/contents/"+contentID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestContentHandler_Update_Success(t *testing.T) {
	mockSvc := new(MockContentService)
	h := NewContentHandler(mockSvc)
	router := setupContentRouter(h)

	contentID := uuid.New()
	sessionID := uuid.New()
	updated := &model.Content{ID: contentID, SessionID: sessionID, Type: model.ContentTypePDF, ContentText: "Updated content"}

	mockSvc.On("Update", mock.Anything, contentID, mock.AnythingOfType("*model.UpdateContentRequest")).Return(updated, nil)

	body, _ := json.Marshal(model.UpdateContentRequest{SessionID: sessionID, Type: model.ContentTypePDF, ContentText: "Updated content"})
	req := httptest.NewRequest(http.MethodPut, "/contents/"+contentID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestContentHandler_Update_InvalidID(t *testing.T) {
	mockSvc := new(MockContentService)
	h := NewContentHandler(mockSvc)
	router := setupContentRouter(h)

	sessionID := uuid.New()
	body, _ := json.Marshal(model.UpdateContentRequest{SessionID: sessionID, Type: model.ContentTypeVideo})
	req := httptest.NewRequest(http.MethodPut, "/contents/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContentHandler_Update_NotFound(t *testing.T) {
	mockSvc := new(MockContentService)
	h := NewContentHandler(mockSvc)
	router := setupContentRouter(h)

	contentID := uuid.New()
	sessionID := uuid.New()
	mockSvc.On("Update", mock.Anything, contentID, mock.AnythingOfType("*model.UpdateContentRequest")).Return(nil, service.ErrContentNotFound)

	body, _ := json.Marshal(model.UpdateContentRequest{SessionID: sessionID, Type: model.ContentTypeVideo})
	req := httptest.NewRequest(http.MethodPut, "/contents/"+contentID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestContentHandler_Delete_Success(t *testing.T) {
	mockSvc := new(MockContentService)
	h := NewContentHandler(mockSvc)
	router := setupContentRouter(h)

	contentID := uuid.New()
	mockSvc.On("Delete", mock.Anything, contentID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/contents/"+contentID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestContentHandler_Delete_InvalidID(t *testing.T) {
	mockSvc := new(MockContentService)
	h := NewContentHandler(mockSvc)
	router := setupContentRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/contents/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContentHandler_Delete_NotFound(t *testing.T) {
	mockSvc := new(MockContentService)
	h := NewContentHandler(mockSvc)
	router := setupContentRouter(h)

	contentID := uuid.New()
	mockSvc.On("Delete", mock.Anything, contentID).Return(service.ErrContentNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/contents/"+contentID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestContentHandler_List_Success(t *testing.T) {
	mockSvc := new(MockContentService)
	h := NewContentHandler(mockSvc)
	router := setupContentRouter(h)

	contents := []*model.Content{
		{ID: uuid.New(), Type: model.ContentTypeVideo},
		{ID: uuid.New(), Type: model.ContentTypePDF},
	}

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListContentsRequest) bool {
		return req.Page == 1 && req.PageSize == 20
	})).Return(contents, int64(2), nil)

	req := httptest.NewRequest(http.MethodGet, "/contents?page=1&page_size=20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestContentHandler_List_WithSearch(t *testing.T) {
	mockSvc := new(MockContentService)
	h := NewContentHandler(mockSvc)
	router := setupContentRouter(h)

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListContentsRequest) bool {
		return req.Type == model.ContentTypeVideo
	})).Return([]*model.Content{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/contents?type=video", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestContentHandler_List_Empty(t *testing.T) {
	mockSvc := new(MockContentService)
	h := NewContentHandler(mockSvc)
	router := setupContentRouter(h)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListContentsRequest")).Return([]*model.Content{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/contents", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestContentHandler_List_ServiceError(t *testing.T) {
	mockSvc := new(MockContentService)
	h := NewContentHandler(mockSvc)
	router := setupContentRouter(h)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListContentsRequest")).Return(([]*model.Content)(nil), int64(0), assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/contents", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func setupContentRouter(handler *ContentHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/contents", handler.Create)
	r.GET("/contents", handler.List)
	r.GET("/contents/:id", handler.Get)
	r.PUT("/contents/:id", handler.Update)
	r.DELETE("/contents/:id", handler.Delete)
	return r
}
