package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/lms/internal/model"
	"github.com/yourusername/lms/pkg/pagination"
	"gorm.io/gorm"
)

type ContentRepository interface {
	Create(ctx context.Context, content *model.Content) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Content, error)
	Update(ctx context.Context, content *model.Content) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *model.ListContentsRequest) ([]*model.Content, int64, error)
}

type contentRepository struct {
	db *gorm.DB
}

func NewContentRepository(db *gorm.DB) ContentRepository {
	return &contentRepository{db: db}
}

func (r *contentRepository) Create(ctx context.Context, content *model.Content) error {
	return r.db.WithContext(ctx).Create(content).Error
}

func (r *contentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Content, error) {
	var content model.Content
	if err := r.db.WithContext(ctx).Preload("Session").Where("id = ?", id).First(&content).Error; err != nil {
		return nil, err
	}
	return &content, nil
}

func (r *contentRepository) Update(ctx context.Context, content *model.Content) error {
	return r.db.WithContext(ctx).Save(content).Error
}

func (r *contentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Content{}, "id = ?", id).Error
}

func (r *contentRepository) List(ctx context.Context, filter *model.ListContentsRequest) ([]*model.Content, int64, error) {
	var contents []*model.Content
	query := r.db.WithContext(ctx).Model(&model.Content{}).Preload("Session")

	if filter.SessionID != uuid.Nil {
		query = query.Where("session_id = ?", filter.SessionID)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}

	total, paginated, err := pagination.Paginate(query, filter.Page, filter.PageSize)
	if err != nil {
		return nil, 0, err
	}

	if err := paginated.Order("created_at DESC").Find(&contents).Error; err != nil {
		return nil, 0, err
	}

	return contents, total, nil
}
