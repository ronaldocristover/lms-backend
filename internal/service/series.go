package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ronaldocristover/lms-backend/internal/model"
	"github.com/ronaldocristover/lms-backend/internal/repository"
	"github.com/ronaldocristover/lms-backend/pkg/apierror"
)

var (
	ErrSeriesNotFound = errors.New("series not found")
)

type SeriesService interface {
	Create(ctx context.Context, req *model.CreateSeriesRequest) (*model.Series, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Series, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateSeriesRequest) (*model.Series, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, req *model.ListSeriesRequest) ([]*model.Series, int64, error)
}

type seriesService struct {
	repo    repository.SeriesRepository
	catRepo repository.CategoryRepository
}

func NewSeriesService(repo repository.SeriesRepository, catRepo repository.CategoryRepository) SeriesService {
	return &seriesService{
		repo:    repo,
		catRepo: catRepo,
	}
}

func (s *seriesService) Create(ctx context.Context, req *model.CreateSeriesRequest) (*model.Series, error) {
	if _, err := s.catRepo.GetByID(ctx, req.CategoryID); err != nil {
		return nil, ErrCategoryNotFound
	}

	series := &model.Series{
		Title:      req.Title,
		CategoryID: req.CategoryID,
		IsPaid:     req.IsPaid,
	}

	if err := s.repo.Create(ctx, series); err != nil {
		return nil, apierror.Internal("Failed to create series")
	}

	return series, nil
}

func (s *seriesService) GetByID(ctx context.Context, id uuid.UUID) (*model.Series, error) {
	series, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrSeriesNotFound
	}
	return series, nil
}

func (s *seriesService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateSeriesRequest) (*model.Series, error) {
	series, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrSeriesNotFound
	}

	if _, err := s.catRepo.GetByID(ctx, req.CategoryID); err != nil {
		return nil, ErrCategoryNotFound
	}

	series.Title = req.Title
	series.CategoryID = req.CategoryID
	series.IsPaid = req.IsPaid

	if err := s.repo.Update(ctx, series); err != nil {
		return nil, apierror.Internal("Failed to update series")
	}

	return series, nil
}

func (s *seriesService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrSeriesNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return apierror.Internal("Failed to delete series")
	}
	return nil
}

func (s *seriesService) List(ctx context.Context, req *model.ListSeriesRequest) ([]*model.Series, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	return s.repo.List(ctx, req)
}
