package service

import (
	"context"
	"testing"
	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ronaldocristover/lms-backend/internal/model"
)

type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) Create(ctx context.Context, category *model.Category) error {
	args := m.Called(ctx, category)
	if args.Error(0) == nil {
		category.ID = uuid.New()
	}
	return args.Error(0)
}

func (m *MockCategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetByName(ctx context.Context, name string) (*model.Category, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Category), args.Error(1)
}

func (m *MockCategoryRepository) Update(ctx context.Context, category *model.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCategoryRepository) List(ctx context.Context, filter *model.ListCategoriesRequest) ([]*model.Category, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Category), args.Get(1).(int64), args.Error(2)
}

func newTestCategoryService() (CategoryService, *MockCategoryRepository) {
	mockRepo := new(MockCategoryRepository)
	return NewCategoryService(mockRepo, zap.NewNop().Sugar()), mockRepo
}

func TestCategoryService_Create_Success(t *testing.T) {
	svc, mockRepo := newTestCategoryService()

	req := &model.CreateCategoryRequest{
		Name: "Programming",
	}

	mockRepo.On("GetByName", mock.Anything, req.Name).Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Category")).Return(nil)

	category, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, category)
	assert.Equal(t, req.Name, category.Name)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_Create_CategoryExists(t *testing.T) {
	svc, mockRepo := newTestCategoryService()

	req := &model.CreateCategoryRequest{
		Name: "Programming",
	}

	existing := &model.Category{ID: uuid.New(), Name: "Programming"}
	mockRepo.On("GetByName", mock.Anything, req.Name).Return(existing, nil)

	category, err := svc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, ErrCategoryExists, err)
	assert.Nil(t, category)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_GetByID_Success(t *testing.T) {
	svc, mockRepo := newTestCategoryService()

	categoryID := uuid.New()
	expected := &model.Category{ID: categoryID, Name: "Programming"}

	mockRepo.On("GetByID", mock.Anything, categoryID).Return(expected, nil)

	category, err := svc.GetByID(context.Background(), categoryID)

	assert.NoError(t, err)
	assert.Equal(t, expected, category)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_GetByID_NotFound(t *testing.T) {
	svc, mockRepo := newTestCategoryService()

	categoryID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, categoryID).Return(nil, assert.AnError)

	category, err := svc.GetByID(context.Background(), categoryID)

	assert.Equal(t, ErrCategoryNotFound, err)
	assert.Nil(t, category)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_Update_Success(t *testing.T) {
	svc, mockRepo := newTestCategoryService()

	categoryID := uuid.New()
	existing := &model.Category{ID: categoryID, Name: "Programming"}

	mockRepo.On("GetByID", mock.Anything, categoryID).Return(existing, nil)
	mockRepo.On("GetByName", mock.Anything, "Web Development").Return(nil, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Category")).Return(nil)

	category, err := svc.Update(context.Background(), categoryID, &model.UpdateCategoryRequest{
		Name: "Web Development",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Web Development", category.Name)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_Update_NotFound(t *testing.T) {
	svc, mockRepo := newTestCategoryService()

	categoryID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, categoryID).Return(nil, assert.AnError)

	category, err := svc.Update(context.Background(), categoryID, &model.UpdateCategoryRequest{Name: "Web Development"})

	assert.Equal(t, ErrCategoryNotFound, err)
	assert.Nil(t, category)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_Update_DuplicateName(t *testing.T) {
	svc, mockRepo := newTestCategoryService()

	categoryID := uuid.New()
	otherID := uuid.New()
	existing := &model.Category{ID: categoryID, Name: "Programming"}
	duplicate := &model.Category{ID: otherID, Name: "Web Development"}

	mockRepo.On("GetByID", mock.Anything, categoryID).Return(existing, nil)
	mockRepo.On("GetByName", mock.Anything, "Web Development").Return(duplicate, nil)

	category, err := svc.Update(context.Background(), categoryID, &model.UpdateCategoryRequest{Name: "Web Development"})

	assert.Equal(t, ErrCategoryExists, err)
	assert.Nil(t, category)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_Delete_Success(t *testing.T) {
	svc, mockRepo := newTestCategoryService()

	categoryID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, categoryID).Return(&model.Category{ID: categoryID}, nil)
	mockRepo.On("Delete", mock.Anything, categoryID).Return(nil)

	err := svc.Delete(context.Background(), categoryID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_Delete_NotFound(t *testing.T) {
	svc, mockRepo := newTestCategoryService()

	categoryID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, categoryID).Return(nil, assert.AnError)

	err := svc.Delete(context.Background(), categoryID)

	assert.Equal(t, ErrCategoryNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_List_Success(t *testing.T) {
	svc, mockRepo := newTestCategoryService()

	categories := []*model.Category{
		{ID: uuid.New(), Name: "Programming"},
		{ID: uuid.New(), Name: "Design"},
	}

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListCategoriesRequest")).Return(categories, int64(2), nil)

	result, total, err := svc.List(context.Background(), &model.ListCategoriesRequest{Page: 1, PageSize: 20})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_List_DefaultPagination(t *testing.T) {
	svc, mockRepo := newTestCategoryService()

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListCategoriesRequest")).Return([]*model.Category{}, int64(0), nil)

	result, total, err := svc.List(context.Background(), &model.ListCategoriesRequest{})

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}

func TestCategoryService_List_WithSearch(t *testing.T) {
	svc, mockRepo := newTestCategoryService()

	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListCategoriesRequest) bool {
		return req.Search == "programming"
	})).Return([]*model.Category{}, int64(0), nil)

	result, _, err := svc.List(context.Background(), &model.ListCategoriesRequest{
		Page:     1,
		PageSize: 20,
		Search:   "programming",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRepo.AssertExpectations(t)
}
