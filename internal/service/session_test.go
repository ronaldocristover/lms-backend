package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yourusername/lms/internal/model"
)

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *model.Session) error {
	args := m.Called(ctx, session)
	if args.Error(0) == nil {
		session.ID = uuid.New()
	}
	return args.Error(0)
}

func (m *MockSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Session), args.Error(1)
}

func (m *MockSessionRepository) Update(ctx context.Context, session *model.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) List(ctx context.Context, filter *model.ListSessionsRequest) ([]*model.Session, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Session), args.Get(1).(int64), args.Error(2)
}

func newTestSessionService() (SessionService, *MockSessionRepository, *MockSeriesRepository) {
	mockSessionRepo := new(MockSessionRepository)
	mockSeriesRepo := new(MockSeriesRepository)
	return NewSessionService(mockSessionRepo, mockSeriesRepo), mockSessionRepo, mockSeriesRepo
}

func TestSessionService_Create_Success(t *testing.T) {
	svc, mockSessionRepo, mockSeriesRepo := newTestSessionService()

	seriesID := uuid.New()
	req := &model.CreateSessionRequest{
		SeriesID:    seriesID,
		Title:       "Introduction to Go",
		Description: "Learn the basics of Go",
		Order:       1,
	}

	mockSeriesRepo.On("GetByID", mock.Anything, seriesID).Return(&model.Series{ID: seriesID}, nil)
	mockSessionRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Session")).Return(nil)

	session, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, req.Title, session.Title)
	assert.Equal(t, req.SeriesID, session.SeriesID)
	mockSessionRepo.AssertExpectations(t)
	mockSeriesRepo.AssertExpectations(t)
}

func TestSessionService_Create_SeriesNotFound(t *testing.T) {
	svc, _, mockSeriesRepo := newTestSessionService()

	seriesID := uuid.New()
	req := &model.CreateSessionRequest{
		SeriesID:    seriesID,
		Title:       "Introduction to Go",
		Description: "Learn the basics of Go",
		Order:       1,
	}

	mockSeriesRepo.On("GetByID", mock.Anything, seriesID).Return(nil, assert.AnError)

	session, err := svc.Create(context.Background(), req)

	assert.Equal(t, ErrSeriesNotFound, err)
	assert.Nil(t, session)
	mockSeriesRepo.AssertExpectations(t)
}

func TestSessionService_GetByID_Success(t *testing.T) {
	svc, mockSessionRepo, _ := newTestSessionService()

	sessionID := uuid.New()
	expected := &model.Session{ID: sessionID, SeriesID: uuid.New(), Title: "Introduction to Go", Description: "Learn the basics", Order: 1}

	mockSessionRepo.On("GetByID", mock.Anything, sessionID).Return(expected, nil)

	session, err := svc.GetByID(context.Background(), sessionID)

	assert.NoError(t, err)
	assert.Equal(t, expected, session)
	mockSessionRepo.AssertExpectations(t)
}

func TestSessionService_GetByID_NotFound(t *testing.T) {
	svc, mockSessionRepo, _ := newTestSessionService()

	sessionID := uuid.New()
	mockSessionRepo.On("GetByID", mock.Anything, sessionID).Return(nil, assert.AnError)

	session, err := svc.GetByID(context.Background(), sessionID)

	assert.Equal(t, ErrSessionNotFound, err)
	assert.Nil(t, session)
	mockSessionRepo.AssertExpectations(t)
}

func TestSessionService_Update_Success(t *testing.T) {
	svc, mockSessionRepo, mockSeriesRepo := newTestSessionService()

	sessionID := uuid.New()
	seriesID := uuid.New()
	existing := &model.Session{ID: sessionID, SeriesID: uuid.New(), Title: "Old Title", Description: "Old Desc", Order: 0}

	mockSessionRepo.On("GetByID", mock.Anything, sessionID).Return(existing, nil)
	mockSeriesRepo.On("GetByID", mock.Anything, seriesID).Return(&model.Series{ID: seriesID}, nil)
	mockSessionRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Session")).Return(nil)

	session, err := svc.Update(context.Background(), sessionID, &model.UpdateSessionRequest{
		SeriesID:    seriesID,
		Title:       "New Title",
		Description: "New Desc",
		Order:       1,
	})

	assert.NoError(t, err)
	assert.Equal(t, "New Title", session.Title)
	assert.Equal(t, seriesID, session.SeriesID)
	assert.Equal(t, "New Desc", session.Description)
	assert.Equal(t, 1, session.Order)
	mockSessionRepo.AssertExpectations(t)
	mockSeriesRepo.AssertExpectations(t)
}

func TestSessionService_Update_NotFound(t *testing.T) {
	svc, mockSessionRepo, _ := newTestSessionService()

	sessionID := uuid.New()
	seriesID := uuid.New()
	mockSessionRepo.On("GetByID", mock.Anything, sessionID).Return(nil, assert.AnError)

	session, err := svc.Update(context.Background(), sessionID, &model.UpdateSessionRequest{
		SeriesID:    seriesID,
		Title:       "New Title",
		Description: "New Desc",
		Order:       1,
	})

	assert.Equal(t, ErrSessionNotFound, err)
	assert.Nil(t, session)
	mockSessionRepo.AssertExpectations(t)
}

func TestSessionService_Update_SeriesNotFound(t *testing.T) {
	svc, mockSessionRepo, mockSeriesRepo := newTestSessionService()

	sessionID := uuid.New()
	seriesID := uuid.New()
	existing := &model.Session{ID: sessionID, SeriesID: uuid.New(), Title: "Old Title", Description: "Old Desc", Order: 0}

	mockSessionRepo.On("GetByID", mock.Anything, sessionID).Return(existing, nil)
	mockSeriesRepo.On("GetByID", mock.Anything, seriesID).Return(nil, assert.AnError)

	session, err := svc.Update(context.Background(), sessionID, &model.UpdateSessionRequest{
		SeriesID:    seriesID,
		Title:       "New Title",
		Description: "New Desc",
		Order:       1,
	})

	assert.Equal(t, ErrSeriesNotFound, err)
	assert.Nil(t, session)
	mockSessionRepo.AssertExpectations(t)
	mockSeriesRepo.AssertExpectations(t)
}

func TestSessionService_Delete_Success(t *testing.T) {
	svc, mockSessionRepo, _ := newTestSessionService()

	sessionID := uuid.New()
	mockSessionRepo.On("GetByID", mock.Anything, sessionID).Return(&model.Session{ID: sessionID}, nil)
	mockSessionRepo.On("Delete", mock.Anything, sessionID).Return(nil)

	err := svc.Delete(context.Background(), sessionID)

	assert.NoError(t, err)
	mockSessionRepo.AssertExpectations(t)
}

func TestSessionService_Delete_NotFound(t *testing.T) {
	svc, mockSessionRepo, _ := newTestSessionService()

	sessionID := uuid.New()
	mockSessionRepo.On("GetByID", mock.Anything, sessionID).Return(nil, assert.AnError)

	err := svc.Delete(context.Background(), sessionID)

	assert.Equal(t, ErrSessionNotFound, err)
	mockSessionRepo.AssertExpectations(t)
}

func TestSessionService_List_Success(t *testing.T) {
	svc, mockSessionRepo, _ := newTestSessionService()

	sessions := []*model.Session{
		{ID: uuid.New(), SeriesID: uuid.New(), Title: "Introduction to Go", Order: 1},
		{ID: uuid.New(), SeriesID: uuid.New(), Title: "Advanced Go", Order: 2},
	}

	mockSessionRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListSessionsRequest")).Return(sessions, int64(2), nil)

	result, total, err := svc.List(context.Background(), &model.ListSessionsRequest{Page: 1, PageSize: 20})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
	mockSessionRepo.AssertExpectations(t)
}

func TestSessionService_List_DefaultPagination(t *testing.T) {
	svc, mockSessionRepo, _ := newTestSessionService()

	mockSessionRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListSessionsRequest")).Return([]*model.Session{}, int64(0), nil)

	result, total, err := svc.List(context.Background(), &model.ListSessionsRequest{})

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, int64(0), total)
	mockSessionRepo.AssertExpectations(t)
}

func TestSessionService_List_WithSearch(t *testing.T) {
	svc, mockSessionRepo, _ := newTestSessionService()

	mockSessionRepo.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListSessionsRequest) bool {
		return req.Search == "introduction"
	})).Return([]*model.Session{}, int64(0), nil)

	result, _, err := svc.List(context.Background(), &model.ListSessionsRequest{
		Page:     1,
		PageSize: 20,
		Search:   "introduction",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockSessionRepo.AssertExpectations(t)
}
