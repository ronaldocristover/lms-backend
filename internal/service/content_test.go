package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yourusername/lms/internal/model"
)

type MockContentRepository struct {
	mock.Mock
}

func (m *MockContentRepository) Create(ctx context.Context, content *model.Content) error {
	args := m.Called(ctx, content)
	if args.Error(0) == nil {
		content.ID = uuid.New()
	}
	return args.Error(0)
}

func (m *MockContentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Content, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Content), args.Error(1)
}

func (m *MockContentRepository) Update(ctx context.Context, content *model.Content) error {
	args := m.Called(ctx, content)
	return args.Error(0)
}

func (m *MockContentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContentRepository) List(ctx context.Context, filter *model.ListContentsRequest) ([]*model.Content, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Content), args.Get(1).(int64), args.Error(2)
}

func newTestContentService() (ContentService, *MockContentRepository, *MockSessionRepository) {
	mockContentRepo := new(MockContentRepository)
	mockSessionRepo := new(MockSessionRepository)
	return NewContentService(mockContentRepo, mockSessionRepo), mockContentRepo, mockSessionRepo
}

func TestContentService_Create_Success(t *testing.T) {
	svc, mockContentRepo, mockSessionRepo := newTestContentService()

	sessionID := uuid.New()
	mediaID := uuid.New()
	req := &model.CreateContentRequest{
		SessionID:   sessionID,
		Type:        model.ContentTypeVideo,
		MediaID:     mediaID,
		ContentText: "Video content",
	}

	mockSessionRepo.On("GetByID", mock.Anything, sessionID).Return(&model.Session{ID: sessionID}, nil)
	mockContentRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Content")).Return(nil)

	content, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, content)
	assert.Equal(t, req.Type, content.Type)
	assert.Equal(t, req.SessionID, content.SessionID)
	mockContentRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestContentService_Create_SessionNotFound(t *testing.T) {
	svc, _, mockSessionRepo := newTestContentService()

	sessionID := uuid.New()
	req := &model.CreateContentRequest{
		SessionID:   sessionID,
		Type:        model.ContentTypeVideo,
		ContentText: "Video content",
	}

	mockSessionRepo.On("GetByID", mock.Anything, sessionID).Return(nil, assert.AnError)

	content, err := svc.Create(context.Background(), req)

	assert.Equal(t, ErrSessionNotFound, err)
	assert.Nil(t, content)
	mockSessionRepo.AssertExpectations(t)
}

func TestContentService_GetByID_Success(t *testing.T) {
	svc, mockContentRepo, _ := newTestContentService()

	contentID := uuid.New()
	expected := &model.Content{ID: contentID, SessionID: uuid.New(), Type: model.ContentTypeVideo, ContentText: "Video content"}

	mockContentRepo.On("GetByID", mock.Anything, contentID).Return(expected, nil)

	content, err := svc.GetByID(context.Background(), contentID)

	assert.NoError(t, err)
	assert.Equal(t, expected, content)
	mockContentRepo.AssertExpectations(t)
}

func TestContentService_GetByID_NotFound(t *testing.T) {
	svc, mockContentRepo, _ := newTestContentService()

	contentID := uuid.New()
	mockContentRepo.On("GetByID", mock.Anything, contentID).Return(nil, assert.AnError)

	content, err := svc.GetByID(context.Background(), contentID)

	assert.Equal(t, ErrContentNotFound, err)
	assert.Nil(t, content)
	mockContentRepo.AssertExpectations(t)
}

func TestContentService_Update_Success(t *testing.T) {
	svc, mockContentRepo, mockSessionRepo := newTestContentService()

	contentID := uuid.New()
	sessionID := uuid.New()
	mediaID := uuid.New()
	existing := &model.Content{ID: contentID, SessionID: uuid.New(), Type: model.ContentTypeText, ContentText: "Old text"}

	mockContentRepo.On("GetByID", mock.Anything, contentID).Return(existing, nil)
	mockSessionRepo.On("GetByID", mock.Anything, sessionID).Return(&model.Session{ID: sessionID}, nil)
	mockContentRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Content")).Return(nil)

	content, err := svc.Update(context.Background(), contentID, &model.UpdateContentRequest{
		SessionID:   sessionID,
		Type:        model.ContentTypeVideo,
		MediaID:     mediaID,
		ContentText: "New text",
	})

	assert.NoError(t, err)
	assert.Equal(t, model.ContentTypeVideo, content.Type)
	assert.Equal(t, sessionID, content.SessionID)
	assert.Equal(t, mediaID, content.MediaID)
	assert.Equal(t, "New text", content.ContentText)
	mockContentRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestContentService_Update_NotFound(t *testing.T) {
	svc, mockContentRepo, _ := newTestContentService()

	contentID := uuid.New()
	sessionID := uuid.New()
	mockContentRepo.On("GetByID", mock.Anything, contentID).Return(nil, assert.AnError)

	content, err := svc.Update(context.Background(), contentID, &model.UpdateContentRequest{
		SessionID:   sessionID,
		Type:        model.ContentTypeVideo,
		ContentText: "New text",
	})

	assert.Equal(t, ErrContentNotFound, err)
	assert.Nil(t, content)
	mockContentRepo.AssertExpectations(t)
}

func TestContentService_Update_SessionNotFound(t *testing.T) {
	svc, mockContentRepo, mockSessionRepo := newTestContentService()

	contentID := uuid.New()
	sessionID := uuid.New()
	existing := &model.Content{ID: contentID, SessionID: uuid.New(), Type: model.ContentTypeText, ContentText: "Old text"}

	mockContentRepo.On("GetByID", mock.Anything, contentID).Return(existing, nil)
	mockSessionRepo.On("GetByID", mock.Anything, sessionID).Return(nil, assert.AnError)

	content, err := svc.Update(context.Background(), contentID, &model.UpdateContentRequest{
		SessionID:   sessionID,
		Type:        model.ContentTypeVideo,
		ContentText: "New text",
	})

	assert.Equal(t, ErrSessionNotFound, err)
	assert.Nil(t, content)
	mockContentRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestContentService_Delete_Success(t *testing.T) {
	svc, mockContentRepo, _ := newTestContentService()

	contentID := uuid.New()
	mockContentRepo.On("GetByID", mock.Anything, contentID).Return(&model.Content{ID: contentID}, nil)
	mockContentRepo.On("Delete", mock.Anything, contentID).Return(nil)

	err := svc.Delete(context.Background(), contentID)

	assert.NoError(t, err)
	mockContentRepo.AssertExpectations(t)
}

func TestContentService_Delete_NotFound(t *testing.T) {
	svc, mockContentRepo, _ := newTestContentService()

	contentID := uuid.New()
	mockContentRepo.On("GetByID", mock.Anything, contentID).Return(nil, assert.AnError)

	err := svc.Delete(context.Background(), contentID)

	assert.Equal(t, ErrContentNotFound, err)
	mockContentRepo.AssertExpectations(t)
}

func TestContentService_List_Success(t *testing.T) {
	svc, mockContentRepo, _ := newTestContentService()

	contents := []*model.Content{
		{ID: uuid.New(), SessionID: uuid.New(), Type: model.ContentTypeVideo},
		{ID: uuid.New(), SessionID: uuid.New(), Type: model.ContentTypeText},
	}

	mockContentRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListContentsRequest")).Return(contents, int64(2), nil)

	result, total, err := svc.List(context.Background(), &model.ListContentsRequest{Page: 1, PageSize: 20})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
	mockContentRepo.AssertExpectations(t)
}

func TestContentService_List_DefaultPagination(t *testing.T) {
	svc, mockContentRepo, _ := newTestContentService()

	mockContentRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListContentsRequest")).Return([]*model.Content{}, int64(0), nil)

	result, total, err := svc.List(context.Background(), &model.ListContentsRequest{})

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, int64(0), total)
	mockContentRepo.AssertExpectations(t)
}

func TestContentService_List_WithSearch(t *testing.T) {
	svc, mockContentRepo, _ := newTestContentService()

	mockContentRepo.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListContentsRequest) bool {
		return req.Type == model.ContentTypeVideo
	})).Return([]*model.Content{}, int64(0), nil)

	result, _, err := svc.List(context.Background(), &model.ListContentsRequest{
		Page:     1,
		PageSize: 20,
		Type:     model.ContentTypeVideo,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockContentRepo.AssertExpectations(t)
}
