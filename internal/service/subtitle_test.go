package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yourusername/lms/internal/model"
	"github.com/yourusername/lms/internal/repository"
)

type MockSubtitleRepository struct {
	mock.Mock
}

func (m *MockSubtitleRepository) Create(ctx context.Context, subtitle *model.Subtitle) error {
	args := m.Called(ctx, subtitle)
	if args.Error(0) == nil {
		subtitle.ID = uuid.New()
	}
	return args.Error(0)
}

func (m *MockSubtitleRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subtitle, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subtitle), args.Error(1)
}

func (m *MockSubtitleRepository) GetByMediaAndLanguage(ctx context.Context, mediaID, languageID uuid.UUID) (*model.Subtitle, error) {
	args := m.Called(ctx, mediaID, languageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subtitle), args.Error(1)
}

func (m *MockSubtitleRepository) Update(ctx context.Context, subtitle *model.Subtitle) error {
	args := m.Called(ctx, subtitle)
	return args.Error(0)
}

func (m *MockSubtitleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSubtitleRepository) List(ctx context.Context, filter *model.ListSubtitlesRequest) ([]*model.Subtitle, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Subtitle), args.Get(1).(int64), args.Error(2)
}

func newTestSubtitleService() (SubtitleService, *MockSubtitleRepository, *MockMediaRepository, *MockLanguageRepository) {
	mockSubtitleRepo := new(MockSubtitleRepository)
	mockMediaRepo := new(MockMediaRepository)
	mockLangRepo := new(MockLanguageRepository)
	return NewSubtitleService(mockSubtitleRepo, mockMediaRepo, mockLangRepo), mockSubtitleRepo, mockMediaRepo, mockLangRepo
}

func TestSubtitleService_Create_Success(t *testing.T) {
	svc, mockRepo, mockMediaRepo, mockLangRepo := newTestSubtitleService()

	mediaID := uuid.New()
	langID := uuid.New()
	req := &model.CreateSubtitleRequest{
		MediaID:    mediaID,
		LanguageID: langID,
		Content:    "Hello world",
	}

	mockMediaRepo.On("GetByID", mock.Anything, mediaID).Return(&model.Media{ID: mediaID}, nil)
	mockLangRepo.On("GetByID", mock.Anything, langID).Return(&model.Language{ID: langID}, nil)
	mockRepo.On("GetByMediaAndLanguage", mock.Anything, mediaID, langID).Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Subtitle")).Return(nil)

	subtitle, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, subtitle)
	assert.Equal(t, mediaID, subtitle.MediaID)
	assert.Equal(t, langID, subtitle.LanguageID)
	assert.Equal(t, "Hello world", subtitle.Content)
	mockRepo.AssertExpectations(t)
	mockMediaRepo.AssertExpectations(t)
	mockLangRepo.AssertExpectations(t)
}

func TestSubtitleService_Create_InvalidMedia(t *testing.T) {
	svc, _, mockMediaRepo, _ := newTestSubtitleService()

	mediaID := uuid.New()
	langID := uuid.New()
	req := &model.CreateSubtitleRequest{
		MediaID:    mediaID,
		LanguageID: langID,
		Content:    "Hello world",
	}

	mockMediaRepo.On("GetByID", mock.Anything, mediaID).Return(nil, assert.AnError)

	subtitle, err := svc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, subtitle)
	mockMediaRepo.AssertExpectations(t)
}

func TestSubtitleService_Create_InvalidLanguage(t *testing.T) {
	svc, _, mockMediaRepo, mockLangRepo := newTestSubtitleService()

	mediaID := uuid.New()
	langID := uuid.New()
	req := &model.CreateSubtitleRequest{
		MediaID:    mediaID,
		LanguageID: langID,
		Content:    "Hello world",
	}

	mockMediaRepo.On("GetByID", mock.Anything, mediaID).Return(&model.Media{ID: mediaID}, nil)
	mockLangRepo.On("GetByID", mock.Anything, langID).Return(nil, assert.AnError)

	subtitle, err := svc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, subtitle)
	mockMediaRepo.AssertExpectations(t)
	mockLangRepo.AssertExpectations(t)
}

func TestSubtitleService_Create_Duplicate(t *testing.T) {
	svc, mockRepo, mockMediaRepo, mockLangRepo := newTestSubtitleService()

	mediaID := uuid.New()
	langID := uuid.New()
	req := &model.CreateSubtitleRequest{
		MediaID:    mediaID,
		LanguageID: langID,
		Content:    "Hello world",
	}

	existing := &model.Subtitle{ID: uuid.New(), MediaID: mediaID, LanguageID: langID, Content: "Existing"}
	mockMediaRepo.On("GetByID", mock.Anything, mediaID).Return(&model.Media{ID: mediaID}, nil)
	mockLangRepo.On("GetByID", mock.Anything, langID).Return(&model.Language{ID: langID}, nil)
	mockRepo.On("GetByMediaAndLanguage", mock.Anything, mediaID, langID).Return(existing, nil)

	subtitle, err := svc.Create(context.Background(), req)

	assert.Equal(t, ErrSubtitleExists, err)
	assert.Nil(t, subtitle)
	mockRepo.AssertExpectations(t)
	mockMediaRepo.AssertExpectations(t)
	mockLangRepo.AssertExpectations(t)
}

func TestSubtitleService_GetByID_Success(t *testing.T) {
	svc, mockRepo, _, _ := newTestSubtitleService()

	subtitleID := uuid.New()
	expected := &model.Subtitle{ID: subtitleID, MediaID: uuid.New(), LanguageID: uuid.New(), Content: "Test"}

	mockRepo.On("GetByID", mock.Anything, subtitleID).Return(expected, nil)

	subtitle, err := svc.GetByID(context.Background(), subtitleID)

	assert.NoError(t, err)
	assert.Equal(t, expected, subtitle)
	mockRepo.AssertExpectations(t)
}

func TestSubtitleService_GetByID_NotFound(t *testing.T) {
	svc, mockRepo, _, _ := newTestSubtitleService()

	subtitleID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, subtitleID).Return(nil, assert.AnError)

	subtitle, err := svc.GetByID(context.Background(), subtitleID)

	assert.Equal(t, ErrSubtitleNotFound, err)
	assert.Nil(t, subtitle)
	mockRepo.AssertExpectations(t)
}

func TestSubtitleService_Update_Success(t *testing.T) {
	svc, mockRepo, _, _ := newTestSubtitleService()

	subtitleID := uuid.New()
	existing := &model.Subtitle{ID: subtitleID, MediaID: uuid.New(), LanguageID: uuid.New(), Content: "Old content"}

	mockRepo.On("GetByID", mock.Anything, subtitleID).Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Subtitle")).Return(nil)

	subtitle, err := svc.Update(context.Background(), subtitleID, &model.UpdateSubtitleRequest{
		Content: "New content",
	})

	assert.NoError(t, err)
	assert.Equal(t, "New content", subtitle.Content)
	mockRepo.AssertExpectations(t)
}

func TestSubtitleService_Update_NotFound(t *testing.T) {
	svc, mockRepo, _, _ := newTestSubtitleService()

	subtitleID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, subtitleID).Return(nil, assert.AnError)

	subtitle, err := svc.Update(context.Background(), subtitleID, &model.UpdateSubtitleRequest{
		Content: "New content",
	})

	assert.Equal(t, ErrSubtitleNotFound, err)
	assert.Nil(t, subtitle)
	mockRepo.AssertExpectations(t)
}

func TestSubtitleService_Delete_Success(t *testing.T) {
	svc, mockRepo, _, _ := newTestSubtitleService()

	subtitleID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, subtitleID).Return(&model.Subtitle{ID: subtitleID}, nil)
	mockRepo.On("Delete", mock.Anything, subtitleID).Return(nil)

	err := svc.Delete(context.Background(), subtitleID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSubtitleService_Delete_NotFound(t *testing.T) {
	svc, mockRepo, _, _ := newTestSubtitleService()

	subtitleID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, subtitleID).Return(nil, assert.AnError)

	err := svc.Delete(context.Background(), subtitleID)

	assert.Equal(t, ErrSubtitleNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestSubtitleService_List_Success(t *testing.T) {
	svc, mockRepo, _, _ := newTestSubtitleService()

	subtitles := []*model.Subtitle{
		{ID: uuid.New(), MediaID: uuid.New(), LanguageID: uuid.New(), Content: "Sub 1"},
		{ID: uuid.New(), MediaID: uuid.New(), LanguageID: uuid.New(), Content: "Sub 2"},
	}

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListSubtitlesRequest")).Return(subtitles, int64(2), nil)

	result, total, err := svc.List(context.Background(), &model.ListSubtitlesRequest{Page: 1, PageSize: 20})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestSubtitleService_List_DefaultPagination(t *testing.T) {
	svc, mockRepo, _, _ := newTestSubtitleService()

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListSubtitlesRequest")).Return([]*model.Subtitle{}, int64(0), nil)

	result, total, err := svc.List(context.Background(), &model.ListSubtitlesRequest{})

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}

func TestSubtitleService_List_WithMediaFilter(t *testing.T) {
	svc, mockRepo, _, _ := newTestSubtitleService()

	mediaID := uuid.New()
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListSubtitlesRequest) bool {
		return req.MediaID == mediaID.String()
	})).Return([]*model.Subtitle{}, int64(0), nil)

	result, _, err := svc.List(context.Background(), &model.ListSubtitlesRequest{
		Page:     1,
		PageSize: 20,
		MediaID:  mediaID.String(),
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestSubtitleService_List_WithLanguageFilter(t *testing.T) {
	svc, mockRepo, _, _ := newTestSubtitleService()

	langID := uuid.New()
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListSubtitlesRequest) bool {
		return req.LanguageID == langID.String()
	})).Return([]*model.Subtitle{}, int64(0), nil)

	result, _, err := svc.List(context.Background(), &model.ListSubtitlesRequest{
		Page:       1,
		PageSize:   20,
		LanguageID: langID.String(),
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestSubtitleService_List_WithSearch(t *testing.T) {
	svc, mockRepo, _, _ := newTestSubtitleService()

	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListSubtitlesRequest) bool {
		return req.Search == "hello"
	})).Return([]*model.Subtitle{}, int64(0), nil)

	result, _, err := svc.List(context.Background(), &model.ListSubtitlesRequest{
		Page:     1,
		PageSize: 20,
		Search:   "hello",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRepo.AssertExpectations(t)
}

var _ repository.SubtitleRepository = (*MockSubtitleRepository)(nil)
