package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ronaldocristover/lms-backend/internal/model"
	"github.com/ronaldocristover/lms-backend/pkg/pagination"
	"gorm.io/gorm"
)

type MediaRepository interface {
	Create(ctx context.Context, media *model.Media) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Media, error)
	Update(ctx context.Context, media *model.Media) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *model.ListMediaRequest) ([]*model.Media, int64, error)
}

type mediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) Create(ctx context.Context, media *model.Media) error {
	return r.db.WithContext(ctx).Create(media).Error
}

func (r *mediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Media, error) {
	var media model.Media
	if err := r.db.WithContext(ctx).Preload("Language").Where("id = ?", id).First(&media).Error; err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *mediaRepository) Update(ctx context.Context, media *model.Media) error {
	return r.db.WithContext(ctx).Save(media).Error
}

func (r *mediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Media{}, "id = ?", id).Error
}

func (r *mediaRepository) List(ctx context.Context, filter *model.ListMediaRequest) ([]*model.Media, int64, error) {
	var medias []*model.Media
	query := r.db.WithContext(ctx).Model(&model.Media{}).Preload("Language")

	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.LanguageID != "" {
		if langID, err := uuid.Parse(filter.LanguageID); err == nil {
			query = query.Where("language_id = ?", langID)
		}
	}
	if filter.Search != "" {
		search := fmt.Sprintf("%%%s%%", filter.Search)
		query = query.Where("url ILIKE ?", search)
	}

	total, paginated, err := pagination.Paginate(query, filter.Page, filter.PageSize)
	if err != nil {
		return nil, 0, err
	}

	if err := paginated.Find(&medias).Error; err != nil {
		return nil, 0, err
	}

	return medias, total, nil
}
