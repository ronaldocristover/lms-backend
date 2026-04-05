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

type MockSeriesRepository struct {
	mock.Mock
}

func (m *MockSeriesRepository) Create(ctx context.Context, series *model.Series) error {
	args := m.Called(ctx, series)
	if args.Error(0) == nil {
		series.ID = uuid.New()
	}
	return args.Error(0)
}

func (m *MockSeriesRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Series, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Series), args.Error(1)
}

func (m *MockSeriesRepository) Update(ctx context.Context, series *model.Series) error {
	args := m.Called(ctx, series)
	return args.Error(0)
}

func (m *MockSeriesRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSeriesRepository) List(ctx context.Context, filter *model.ListSeriesRequest) ([]*model.Series, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Series), args.Get(1).(int64), args.Error(2)
}

func newTestSeriesService() (SeriesService, *MockSeriesRepository, *MockCategoryRepository) {
	mockSeriesRepo := new(MockSeriesRepository)
	mockCatRepo := new(MockCategoryRepository)
	return NewSeriesService(mockSeriesRepo, mockCatRepo, zap.NewNop().Sugar()), mockSeriesRepo, mockCatRepo
}

func TestSeriesService_Create_Success(t *testing.T) {
	svc, mockSeriesRepo, mockCatRepo := newTestSeriesService()

	categoryID := uuid.New()
	req := &model.CreateSeriesRequest{
		Title:      "Go Programming",
		CategoryID: categoryID,
		IsPaid:     true,
	}

	mockCatRepo.On("GetByID", mock.Anything, categoryID).Return(&model.Category{ID: categoryID}, nil)
	mockSeriesRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Series")).Return(nil)

	series, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, series)
	assert.Equal(t, req.Title, series.Title)
	assert.Equal(t, req.CategoryID, series.CategoryID)
	mockSeriesRepo.AssertExpectations(t)
	mockCatRepo.AssertExpectations(t)
}

func TestSeriesService_Create_CategoryNotFound(t *testing.T) {
	svc, _, mockCatRepo := newTestSeriesService()

	categoryID := uuid.New()
	req := &model.CreateSeriesRequest{
		Title:      "Go Programming",
		CategoryID: categoryID,
	}

	mockCatRepo.On("GetByID", mock.Anything, categoryID).Return(nil, assert.AnError)

	series, err := svc.Create(context.Background(), req)

	assert.Equal(t, ErrCategoryNotFound, err)
	assert.Nil(t, series)
	mockCatRepo.AssertExpectations(t)
}

func TestSeriesService_GetByID_Success(t *testing.T) {
	svc, mockSeriesRepo, _ := newTestSeriesService()

	seriesID := uuid.New()
	expected := &model.Series{ID: seriesID, Title: "Go Programming", CategoryID: uuid.New(), IsPaid: true}

	mockSeriesRepo.On("GetByID", mock.Anything, seriesID).Return(expected, nil)

	series, err := svc.GetByID(context.Background(), seriesID)

	assert.NoError(t, err)
	assert.Equal(t, expected, series)
	mockSeriesRepo.AssertExpectations(t)
}

func TestSeriesService_GetByID_NotFound(t *testing.T) {
	svc, mockSeriesRepo, _ := newTestSeriesService()

	seriesID := uuid.New()
	mockSeriesRepo.On("GetByID", mock.Anything, seriesID).Return(nil, assert.AnError)

	series, err := svc.GetByID(context.Background(), seriesID)

	assert.Equal(t, ErrSeriesNotFound, err)
	assert.Nil(t, series)
	mockSeriesRepo.AssertExpectations(t)
}

func TestSeriesService_Update_Success(t *testing.T) {
	svc, mockSeriesRepo, mockCatRepo := newTestSeriesService()

	seriesID := uuid.New()
	categoryID := uuid.New()
	existing := &model.Series{ID: seriesID, Title: "Old Title", CategoryID: uuid.New(), IsPaid: false}

	mockSeriesRepo.On("GetByID", mock.Anything, seriesID).Return(existing, nil)
	mockCatRepo.On("GetByID", mock.Anything, categoryID).Return(&model.Category{ID: categoryID}, nil)
	mockSeriesRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Series")).Return(nil)

	series, err := svc.Update(context.Background(), seriesID, &model.UpdateSeriesRequest{
		Title:      "New Title",
		CategoryID: categoryID,
		IsPaid:     true,
	})

	assert.NoError(t, err)
	assert.Equal(t, "New Title", series.Title)
	assert.Equal(t, categoryID, series.CategoryID)
	assert.True(t, series.IsPaid)
	mockSeriesRepo.AssertExpectations(t)
	mockCatRepo.AssertExpectations(t)
}

func TestSeriesService_Update_NotFound(t *testing.T) {
	svc, mockSeriesRepo, _ := newTestSeriesService()

	seriesID := uuid.New()
	categoryID := uuid.New()
	mockSeriesRepo.On("GetByID", mock.Anything, seriesID).Return(nil, assert.AnError)

	series, err := svc.Update(context.Background(), seriesID, &model.UpdateSeriesRequest{
		Title:      "New Title",
		CategoryID: categoryID,
		IsPaid:     true,
	})

	assert.Equal(t, ErrSeriesNotFound, err)
	assert.Nil(t, series)
	mockSeriesRepo.AssertExpectations(t)
}

func TestSeriesService_Update_CategoryNotFound(t *testing.T) {
	svc, mockSeriesRepo, mockCatRepo := newTestSeriesService()

	seriesID := uuid.New()
	categoryID := uuid.New()
	existing := &model.Series{ID: seriesID, Title: "Old Title", CategoryID: uuid.New(), IsPaid: false}

	mockSeriesRepo.On("GetByID", mock.Anything, seriesID).Return(existing, nil)
	mockCatRepo.On("GetByID", mock.Anything, categoryID).Return(nil, assert.AnError)

	series, err := svc.Update(context.Background(), seriesID, &model.UpdateSeriesRequest{
		Title:      "New Title",
		CategoryID: categoryID,
		IsPaid:     true,
	})

	assert.Equal(t, ErrCategoryNotFound, err)
	assert.Nil(t, series)
	mockSeriesRepo.AssertExpectations(t)
	mockCatRepo.AssertExpectations(t)
}

func TestSeriesService_Delete_Success(t *testing.T) {
	svc, mockSeriesRepo, _ := newTestSeriesService()

	seriesID := uuid.New()
	mockSeriesRepo.On("GetByID", mock.Anything, seriesID).Return(&model.Series{ID: seriesID}, nil)
	mockSeriesRepo.On("Delete", mock.Anything, seriesID).Return(nil)

	err := svc.Delete(context.Background(), seriesID)

	assert.NoError(t, err)
	mockSeriesRepo.AssertExpectations(t)
}

func TestSeriesService_Delete_NotFound(t *testing.T) {
	svc, mockSeriesRepo, _ := newTestSeriesService()

	seriesID := uuid.New()
	mockSeriesRepo.On("GetByID", mock.Anything, seriesID).Return(nil, assert.AnError)

	err := svc.Delete(context.Background(), seriesID)

	assert.Equal(t, ErrSeriesNotFound, err)
	mockSeriesRepo.AssertExpectations(t)
}

func TestSeriesService_List_Success(t *testing.T) {
	svc, mockSeriesRepo, _ := newTestSeriesService()

	series := []*model.Series{
		{ID: uuid.New(), Title: "Go Programming", CategoryID: uuid.New(), IsPaid: true},
		{ID: uuid.New(), Title: "Python Basics", CategoryID: uuid.New(), IsPaid: false},
	}

	mockSeriesRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListSeriesRequest")).Return(series, int64(2), nil)

	result, total, err := svc.List(context.Background(), &model.ListSeriesRequest{Page: 1, PageSize: 20})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
	mockSeriesRepo.AssertExpectations(t)
}

func TestSeriesService_List_DefaultPagination(t *testing.T) {
	svc, mockSeriesRepo, _ := newTestSeriesService()

	mockSeriesRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListSeriesRequest")).Return([]*model.Series{}, int64(0), nil)

	result, total, err := svc.List(context.Background(), &model.ListSeriesRequest{})

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, int64(0), total)
	mockSeriesRepo.AssertExpectations(t)
}

func TestSeriesService_List_WithSearch(t *testing.T) {
	svc, mockSeriesRepo, _ := newTestSeriesService()

	mockSeriesRepo.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListSeriesRequest) bool {
		return req.Search == "go"
	})).Return([]*model.Series{}, int64(0), nil)

	result, _, err := svc.List(context.Background(), &model.ListSeriesRequest{
		Page:     1,
		PageSize: 20,
		Search:   "go",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockSeriesRepo.AssertExpectations(t)
}
