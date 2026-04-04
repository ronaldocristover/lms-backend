package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/lms/internal/model"
	"gorm.io/gorm"
)

type SeriesRepository interface {
	Create(ctx context.Context, series *model.Series) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Series, error)
	Update(ctx context.Context, series *model.Series) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *model.ListSeriesRequest) ([]*model.Series, int64, error)
}

type seriesRepository struct {
	db *gorm.DB
}

func NewSeriesRepository(db *gorm.DB) SeriesRepository {
	return &seriesRepository{db: db}
}

func (r *seriesRepository) Create(ctx context.Context, series *model.Series) error {
	return r.db.WithContext(ctx).Create(series).Error
}

func (r *seriesRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Series, error) {
	var series model.Series
	if err := r.db.WithContext(ctx).Preload("Category").Where("id = ?", id).First(&series).Error; err != nil {
		return nil, err
	}
	return &series, nil
}

func (r *seriesRepository) Update(ctx context.Context, series *model.Series) error {
	return r.db.WithContext(ctx).Save(series).Error
}

func (r *seriesRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Series{}, "id = ?", id).Error
}

func (r *seriesRepository) List(ctx context.Context, filter *model.ListSeriesRequest) ([]*model.Series, int64, error) {
	var series []*model.Series
	var total int64

	page := filter.Page
	pageSize := filter.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := r.db.WithContext(ctx).Model(&model.Series{}).Preload("Category")

	if filter.Search != "" {
		search := fmt.Sprintf("%%%s%%", filter.Search)
		query = query.Where("title ILIKE ?", search)
	}

	if filter.CategoryID != uuid.Nil {
		query = query.Where("category_id = ?", filter.CategoryID)
	}

	if filter.IsPaid != nil {
		query = query.Where("is_paid = ?", *filter.IsPaid)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&series).Error; err != nil {
		return nil, 0, err
	}

	return series, total, nil
}
