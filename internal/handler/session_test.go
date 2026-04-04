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

type MockSessionService struct {
	mock.Mock
}

func (m *MockSessionService) Create(ctx context.Context, req *model.CreateSessionRequest) (*model.Session, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Session), args.Error(1)
}

func (m *MockSessionService) GetByID(ctx context.Context, id uuid.UUID) (*model.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Session), args.Error(1)
}

func (m *MockSessionService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateSessionRequest) (*model.Session, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Session), args.Error(1)
}

func (m *MockSessionService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionService) List(ctx context.Context, req *model.ListSessionsRequest) ([]*model.Session, int64, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Session), args.Get(1).(int64), args.Error(2)
}

func TestSessionHandler_Create_Success(t *testing.T) {
	mockSvc := new(MockSessionService)
	h := NewSessionHandler(mockSvc)
	router := setupSessionRouter(h)

	sessionID := uuid.New()
	seriesID := uuid.New()
	created := &model.Session{ID: sessionID, SeriesID: seriesID, Title: "Introduction", Description: "Welcome", Order: 1}

	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateSessionRequest")).Return(created, nil)

	body, _ := json.Marshal(model.CreateSessionRequest{SeriesID: seriesID, Title: "Introduction", Description: "Welcome", Order: 1})
	req := httptest.NewRequest(http.MethodPost, "/sessions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_Create_SeriesNotFound(t *testing.T) {
	mockSvc := new(MockSessionService)
	h := NewSessionHandler(mockSvc)
	router := setupSessionRouter(h)

	seriesID := uuid.New()
	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateSessionRequest")).Return(nil, service.ErrSeriesNotFound)

	body, _ := json.Marshal(model.CreateSessionRequest{SeriesID: seriesID, Title: "Introduction"})
	req := httptest.NewRequest(http.MethodPost, "/sessions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_Create_InvalidInput(t *testing.T) {
	mockSvc := new(MockSessionService)
	h := NewSessionHandler(mockSvc)
	router := setupSessionRouter(h)

	body, _ := json.Marshal(map[string]string{"invalid": "data"})
	req := httptest.NewRequest(http.MethodPost, "/sessions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSessionHandler_Get_Success(t *testing.T) {
	mockSvc := new(MockSessionService)
	h := NewSessionHandler(mockSvc)
	router := setupSessionRouter(h)

	sessionID := uuid.New()
	session := &model.Session{ID: sessionID, Title: "Introduction"}

	mockSvc.On("GetByID", mock.Anything, sessionID).Return(session, nil)

	req := httptest.NewRequest(http.MethodGet, "/sessions/"+sessionID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_Get_InvalidID(t *testing.T) {
	mockSvc := new(MockSessionService)
	h := NewSessionHandler(mockSvc)
	router := setupSessionRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/sessions/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSessionHandler_Get_NotFound(t *testing.T) {
	mockSvc := new(MockSessionService)
	h := NewSessionHandler(mockSvc)
	router := setupSessionRouter(h)

	sessionID := uuid.New()
	mockSvc.On("GetByID", mock.Anything, sessionID).Return(nil, service.ErrSessionNotFound)

	req := httptest.NewRequest(http.MethodGet, "/sessions/"+sessionID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_Update_Success(t *testing.T) {
	mockSvc := new(MockSessionService)
	h := NewSessionHandler(mockSvc)
	router := setupSessionRouter(h)

	sessionID := uuid.New()
	seriesID := uuid.New()
	updated := &model.Session{ID: sessionID, SeriesID: seriesID, Title: "Updated Session", Description: "Updated", Order: 2}

	mockSvc.On("Update", mock.Anything, sessionID, mock.AnythingOfType("*model.UpdateSessionRequest")).Return(updated, nil)

	body, _ := json.Marshal(model.UpdateSessionRequest{SeriesID: seriesID, Title: "Updated Session", Description: "Updated", Order: 2})
	req := httptest.NewRequest(http.MethodPut, "/sessions/"+sessionID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_Update_InvalidID(t *testing.T) {
	mockSvc := new(MockSessionService)
	h := NewSessionHandler(mockSvc)
	router := setupSessionRouter(h)

	seriesID := uuid.New()
	body, _ := json.Marshal(model.UpdateSessionRequest{SeriesID: seriesID, Title: "Updated Session"})
	req := httptest.NewRequest(http.MethodPut, "/sessions/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSessionHandler_Update_NotFound(t *testing.T) {
	mockSvc := new(MockSessionService)
	h := NewSessionHandler(mockSvc)
	router := setupSessionRouter(h)

	sessionID := uuid.New()
	seriesID := uuid.New()
	mockSvc.On("Update", mock.Anything, sessionID, mock.AnythingOfType("*model.UpdateSessionRequest")).Return(nil, service.ErrSessionNotFound)

	body, _ := json.Marshal(model.UpdateSessionRequest{SeriesID: seriesID, Title: "Updated Session"})
	req := httptest.NewRequest(http.MethodPut, "/sessions/"+sessionID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_Delete_Success(t *testing.T) {
	mockSvc := new(MockSessionService)
	h := NewSessionHandler(mockSvc)
	router := setupSessionRouter(h)

	sessionID := uuid.New()
	mockSvc.On("Delete", mock.Anything, sessionID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/sessions/"+sessionID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_Delete_InvalidID(t *testing.T) {
	mockSvc := new(MockSessionService)
	h := NewSessionHandler(mockSvc)
	router := setupSessionRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/sessions/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSessionHandler_Delete_NotFound(t *testing.T) {
	mockSvc := new(MockSessionService)
	h := NewSessionHandler(mockSvc)
	router := setupSessionRouter(h)

	sessionID := uuid.New()
	mockSvc.On("Delete", mock.Anything, sessionID).Return(service.ErrSessionNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/sessions/"+sessionID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_List_Success(t *testing.T) {
	mockSvc := new(MockSessionService)
	h := NewSessionHandler(mockSvc)
	router := setupSessionRouter(h)

	sessions := []*model.Session{
		{ID: uuid.New(), Title: "Introduction"},
		{ID: uuid.New(), Title: "Advanced Topics"},
	}

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListSessionsRequest) bool {
		return req.Page == 1 && req.PageSize == 20
	})).Return(sessions, int64(2), nil)

	req := httptest.NewRequest(http.MethodGet, "/sessions?page=1&page_size=20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_List_WithSearch(t *testing.T) {
	mockSvc := new(MockSessionService)
	h := NewSessionHandler(mockSvc)
	router := setupSessionRouter(h)

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListSessionsRequest) bool {
		return req.Search == "intro"
	})).Return([]*model.Session{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/sessions?search=intro", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_List_Empty(t *testing.T) {
	mockSvc := new(MockSessionService)
	h := NewSessionHandler(mockSvc)
	router := setupSessionRouter(h)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListSessionsRequest")).Return([]*model.Session{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/sessions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_List_ServiceError(t *testing.T) {
	mockSvc := new(MockSessionService)
	h := NewSessionHandler(mockSvc)
	router := setupSessionRouter(h)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListSessionsRequest")).Return(([]*model.Session)(nil), int64(0), assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/sessions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func setupSessionRouter(handler *SessionHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/sessions", handler.Create)
	r.GET("/sessions", handler.List)
	r.GET("/sessions/:id", handler.Get)
	r.PUT("/sessions/:id", handler.Update)
	r.DELETE("/sessions/:id", handler.Delete)
	return r
}
