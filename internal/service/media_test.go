package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yourusername/lms/internal/model"
)

type MockMediaRepository struct {
	mock.Mock
}

func (m *MockMediaRepository) Create(ctx context.Context, media *model.Media) error {
	args := m.Called(ctx, media)
	if args.Error(0) == nil {
		media.ID = uuid.New()
	}
	return args.Error(0)
}

func (m *MockMediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Media, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Media), args.Error(1)
}

func (m *MockMediaRepository) Update(ctx context.Context, media *model.Media) error {
	args := m.Called(ctx, media)
	return args.Error(0)
}

func (m *MockMediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMediaRepository) List(ctx context.Context, filter *model.ListMediaRequest) ([]*model.Media, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Media), args.Get(1).(int64), args.Error(2)
}

func newTestMediaService() (MediaService, *MockMediaRepository, *MockLanguageRepository) {
	mockRepo := new(MockMediaRepository)
	mockLangRepo := new(MockLanguageRepository)
	return NewMediaService(mockRepo, mockLangRepo), mockRepo, mockLangRepo
}

func TestMediaService_Create_Success(t *testing.T) {
	svc, mockRepo, mockLangRepo := newTestMediaService()

	req := &model.CreateMediaRequest{
		Type: model.MediaTypeVideo,
		URL:  "https://example.com/video.mp4",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Media")).Return(nil)

	media, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, media)
	assert.Equal(t, req.Type, media.Type)
	assert.Equal(t, req.URL, media.URL)
	mockRepo.AssertExpectations(t)
	mockLangRepo.AssertExpectations(t)
}

func TestMediaService_Create_WithLanguage(t *testing.T) {
	svc, mockRepo, mockLangRepo := newTestMediaService()

	langID := uuid.New()
	req := &model.CreateMediaRequest{
		Type:       model.MediaTypeVideo,
		URL:        "https://example.com/video.mp4",
		LanguageID: &langID,
	}

	mockLangRepo.On("GetByID", mock.Anything, langID).Return(&model.Language{ID: langID, Code: "en"}, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Media")).Return(nil)

	media, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, media)
	assert.Equal(t, langID, *media.LanguageID)
	mockRepo.AssertExpectations(t)
	mockLangRepo.AssertExpectations(t)
}

func TestMediaService_Create_InvalidLanguage(t *testing.T) {
	svc, mockRepo, mockLangRepo := newTestMediaService()

	langID := uuid.New()
	req := &model.CreateMediaRequest{
		Type:       model.MediaTypeVideo,
		URL:        "https://example.com/video.mp4",
		LanguageID: &langID,
	}

	mockLangRepo.On("GetByID", mock.Anything, langID).Return(nil, assert.AnError)

	media, err := svc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, media)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	mockLangRepo.AssertExpectations(t)
}

func TestMediaService_GetByID_Success(t *testing.T) {
	svc, mockRepo, _ := newTestMediaService()

	mediaID := uuid.New()
	expected := &model.Media{ID: mediaID, Type: model.MediaTypeVideo, URL: "https://example.com/video.mp4"}

	mockRepo.On("GetByID", mock.Anything, mediaID).Return(expected, nil)

	media, err := svc.GetByID(context.Background(), mediaID)

	assert.NoError(t, err)
	assert.Equal(t, expected, media)
	mockRepo.AssertExpectations(t)
}

func TestMediaService_GetByID_NotFound(t *testing.T) {
	svc, mockRepo, _ := newTestMediaService()

	mediaID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, mediaID).Return(nil, assert.AnError)

	media, err := svc.GetByID(context.Background(), mediaID)

	assert.Equal(t, ErrMediaNotFound, err)
	assert.Nil(t, media)
	mockRepo.AssertExpectations(t)
}

func TestMediaService_Update_Success(t *testing.T) {
	svc, mockRepo, _ := newTestMediaService()

	mediaID := uuid.New()
	existing := &model.Media{ID: mediaID, Type: model.MediaTypeVideo, URL: "https://example.com/old.mp4"}

	mockRepo.On("GetByID", mock.Anything, mediaID).Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Media")).Return(nil)

	media, err := svc.Update(context.Background(), mediaID, &model.UpdateMediaRequest{
		URL: "https://example.com/new.mp4",
	})

	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/new.mp4", media.URL)
	mockRepo.AssertExpectations(t)
}

func TestMediaService_Update_NotFound(t *testing.T) {
	svc, mockRepo, _ := newTestMediaService()

	mediaID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, mediaID).Return(nil, assert.AnError)

	media, err := svc.Update(context.Background(), mediaID, &model.UpdateMediaRequest{
		URL: "https://example.com/new.mp4",
	})

	assert.Equal(t, ErrMediaNotFound, err)
	assert.Nil(t, media)
	mockRepo.AssertExpectations(t)
}

func TestMediaService_Update_InvalidLanguage(t *testing.T) {
	svc, mockRepo, mockLangRepo := newTestMediaService()

	mediaID := uuid.New()
	langID := uuid.New()
	existing := &model.Media{ID: mediaID, Type: model.MediaTypeVideo, URL: "https://example.com/video.mp4"}

	mockRepo.On("GetByID", mock.Anything, mediaID).Return(existing, nil)
	mockLangRepo.On("GetByID", mock.Anything, langID).Return(nil, assert.AnError)

	media, err := svc.Update(context.Background(), mediaID, &model.UpdateMediaRequest{
		LanguageID: &langID,
	})

	assert.Error(t, err)
	assert.Nil(t, media)
	mockRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	mockLangRepo.AssertExpectations(t)
}

func TestMediaService_Delete_Success(t *testing.T) {
	svc, mockRepo, _ := newTestMediaService()

	mediaID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, mediaID).Return(&model.Media{ID: mediaID}, nil)
	mockRepo.On("Delete", mock.Anything, mediaID).Return(nil)

	err := svc.Delete(context.Background(), mediaID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMediaService_Delete_NotFound(t *testing.T) {
	svc, mockRepo, _ := newTestMediaService()

	mediaID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, mediaID).Return(nil, assert.AnError)

	err := svc.Delete(context.Background(), mediaID)

	assert.Equal(t, ErrMediaNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestMediaService_List_Success(t *testing.T) {
	svc, mockRepo, _ := newTestMediaService()

	medias := []*model.Media{
		{ID: uuid.New(), Type: model.MediaTypeVideo, URL: "https://example.com/video1.mp4"},
		{ID: uuid.New(), Type: model.MediaTypeAudio, URL: "https://example.com/audio1.mp3"},
	}

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListMediaRequest")).Return(medias, int64(2), nil)

	result, total, err := svc.List(context.Background(), &model.ListMediaRequest{Page: 1, PageSize: 20})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestMediaService_List_DefaultPagination(t *testing.T) {
	svc, mockRepo, _ := newTestMediaService()

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListMediaRequest")).Return([]*model.Media{}, int64(0), nil)

	result, total, err := svc.List(context.Background(), &model.ListMediaRequest{})

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}

func TestMediaService_List_WithType(t *testing.T) {
	svc, mockRepo, _ := newTestMediaService()

	medias := []*model.Media{
		{ID: uuid.New(), Type: model.MediaTypeVideo, URL: "https://example.com/video1.mp4"},
	}

	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListMediaRequest) bool {
		return req.Type == model.MediaTypeVideo
	})).Return(medias, int64(1), nil)

	result, total, err := svc.List(context.Background(), &model.ListMediaRequest{
		Page:     1,
		PageSize: 20,
		Type:     model.MediaTypeVideo,
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, int64(1), total)
	mockRepo.AssertExpectations(t)
}

func TestMediaService_List_WithSearch(t *testing.T) {
	svc, mockRepo, _ := newTestMediaService()

	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListMediaRequest) bool {
		return req.Search == "video"
	})).Return([]*model.Media{}, int64(0), nil)

	result, _, err := svc.List(context.Background(), &model.ListMediaRequest{
		Page:     1,
		PageSize: 20,
		Search:   "video",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRepo.AssertExpectations(t)
}
