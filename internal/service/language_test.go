package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yourusername/lms/internal/model"
)

type MockLanguageRepository struct {
	mock.Mock
}

func (m *MockLanguageRepository) Create(ctx context.Context, language *model.Language) error {
	args := m.Called(ctx, language)
	if args.Error(0) == nil {
		language.ID = uuid.New()
	}
	return args.Error(0)
}

func (m *MockLanguageRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Language, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Language), args.Error(1)
}

func (m *MockLanguageRepository) GetByCode(ctx context.Context, code string) (*model.Language, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Language), args.Error(1)
}

func (m *MockLanguageRepository) Update(ctx context.Context, language *model.Language) error {
	args := m.Called(ctx, language)
	return args.Error(0)
}

func (m *MockLanguageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockLanguageRepository) List(ctx context.Context, filter *model.ListLanguagesRequest) ([]*model.Language, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Language), args.Get(1).(int64), args.Error(2)
}

func newTestLanguageService() (LanguageService, *MockLanguageRepository) {
	mockRepo := new(MockLanguageRepository)
	return NewLanguageService(mockRepo), mockRepo
}

func TestLanguageService_Create_Success(t *testing.T) {
	svc, mockRepo := newTestLanguageService()

	req := &model.CreateLanguageRequest{
		Code: "en",
		Name: "English",
	}

	mockRepo.On("GetByCode", mock.Anything, req.Code).Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Language")).Return(nil)

	language, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, language)
	assert.Equal(t, req.Code, language.Code)
	assert.Equal(t, req.Name, language.Name)
	mockRepo.AssertExpectations(t)
}

func TestLanguageService_Create_LanguageExists(t *testing.T) {
	svc, mockRepo := newTestLanguageService()

	req := &model.CreateLanguageRequest{
		Code: "en",
		Name: "English",
	}

	existing := &model.Language{ID: uuid.New(), Code: "en", Name: "English"}
	mockRepo.On("GetByCode", mock.Anything, req.Code).Return(existing, nil)

	language, err := svc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, ErrLanguageExists, err)
	assert.Nil(t, language)
	mockRepo.AssertExpectations(t)
}

func TestLanguageService_GetByID_Success(t *testing.T) {
	svc, mockRepo := newTestLanguageService()

	languageID := uuid.New()
	expected := &model.Language{ID: languageID, Code: "en", Name: "English"}

	mockRepo.On("GetByID", mock.Anything, languageID).Return(expected, nil)

	language, err := svc.GetByID(context.Background(), languageID)

	assert.NoError(t, err)
	assert.Equal(t, expected, language)
	mockRepo.AssertExpectations(t)
}

func TestLanguageService_GetByID_NotFound(t *testing.T) {
	svc, mockRepo := newTestLanguageService()

	languageID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, languageID).Return(nil, assert.AnError)

	language, err := svc.GetByID(context.Background(), languageID)

	assert.Equal(t, ErrLanguageNotFound, err)
	assert.Nil(t, language)
	mockRepo.AssertExpectations(t)
}

func TestLanguageService_Update_Success(t *testing.T) {
	svc, mockRepo := newTestLanguageService()

	languageID := uuid.New()
	existing := &model.Language{ID: languageID, Code: "en", Name: "English"}

	mockRepo.On("GetByID", mock.Anything, languageID).Return(existing, nil)
	mockRepo.On("GetByCode", mock.Anything, "es").Return(nil, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Language")).Return(nil)

	language, err := svc.Update(context.Background(), languageID, &model.UpdateLanguageRequest{
		Code: "es",
		Name: "Spanish",
	})

	assert.NoError(t, err)
	assert.Equal(t, "es", language.Code)
	assert.Equal(t, "Spanish", language.Name)
	mockRepo.AssertExpectations(t)
}

func TestLanguageService_Update_NotFound(t *testing.T) {
	svc, mockRepo := newTestLanguageService()

	languageID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, languageID).Return(nil, assert.AnError)

	language, err := svc.Update(context.Background(), languageID, &model.UpdateLanguageRequest{Code: "es"})

	assert.Equal(t, ErrLanguageNotFound, err)
	assert.Nil(t, language)
	mockRepo.AssertExpectations(t)
}

func TestLanguageService_Update_DuplicateCode(t *testing.T) {
	svc, mockRepo := newTestLanguageService()

	languageID := uuid.New()
	otherID := uuid.New()
	existing := &model.Language{ID: languageID, Code: "en", Name: "English"}
	duplicate := &model.Language{ID: otherID, Code: "es", Name: "Spanish"}

	mockRepo.On("GetByID", mock.Anything, languageID).Return(existing, nil)
	mockRepo.On("GetByCode", mock.Anything, "es").Return(duplicate, nil)

	language, err := svc.Update(context.Background(), languageID, &model.UpdateLanguageRequest{Code: "es"})

	assert.Equal(t, ErrLanguageExists, err)
	assert.Nil(t, language)
	mockRepo.AssertExpectations(t)
}

func TestLanguageService_Delete_Success(t *testing.T) {
	svc, mockRepo := newTestLanguageService()

	languageID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, languageID).Return(&model.Language{ID: languageID}, nil)
	mockRepo.On("Delete", mock.Anything, languageID).Return(nil)

	err := svc.Delete(context.Background(), languageID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLanguageService_Delete_NotFound(t *testing.T) {
	svc, mockRepo := newTestLanguageService()

	languageID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, languageID).Return(nil, assert.AnError)

	err := svc.Delete(context.Background(), languageID)

	assert.Equal(t, ErrLanguageNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestLanguageService_List_Success(t *testing.T) {
	svc, mockRepo := newTestLanguageService()

	languages := []*model.Language{
		{ID: uuid.New(), Code: "en", Name: "English"},
		{ID: uuid.New(), Code: "es", Name: "Spanish"},
	}

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListLanguagesRequest")).Return(languages, int64(2), nil)

	result, total, err := svc.List(context.Background(), &model.ListLanguagesRequest{Page: 1, PageSize: 20})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestLanguageService_List_DefaultPagination(t *testing.T) {
	svc, mockRepo := newTestLanguageService()

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListLanguagesRequest")).Return([]*model.Language{}, int64(0), nil)

	result, total, err := svc.List(context.Background(), &model.ListLanguagesRequest{})

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}

func TestLanguageService_List_WithSearch(t *testing.T) {
	svc, mockRepo := newTestLanguageService()

	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListLanguagesRequest) bool {
		return req.Search == "en"
	})).Return([]*model.Language{}, int64(0), nil)

	result, _, err := svc.List(context.Background(), &model.ListLanguagesRequest{
		Page:     1,
		PageSize: 20,
		Search:   "en",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRepo.AssertExpectations(t)
}
