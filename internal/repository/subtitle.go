package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/lms/internal/model"
	"gorm.io/gorm"
)

type SubtitleRepository interface {
	Create(ctx context.Context, subtitle *model.Subtitle) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subtitle, error)
	GetByMediaAndLanguage(ctx context.Context, mediaID, languageID uuid.UUID) (*model.Subtitle, error)
	Update(ctx context.Context, subtitle *model.Subtitle) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *model.ListSubtitlesRequest) ([]*model.Subtitle, int64, error)
}

type subtitleRepository struct {
	db *gorm.DB
}

func NewSubtitleRepository(db *gorm.DB) SubtitleRepository {
	return &subtitleRepository{db: db}
}

func (r *subtitleRepository) Create(ctx context.Context, subtitle *model.Subtitle) error {
	return r.db.WithContext(ctx).Create(subtitle).Error
}

func (r *subtitleRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subtitle, error) {
	var subtitle model.Subtitle
	if err := r.db.WithContext(ctx).Preload("Media").Preload("Media.Language").Preload("Language").Where("id = ?", id).First(&subtitle).Error; err != nil {
		return nil, err
	}
	return &subtitle, nil
}

func (r *subtitleRepository) GetByMediaAndLanguage(ctx context.Context, mediaID, languageID uuid.UUID) (*model.Subtitle, error) {
	var subtitle model.Subtitle
	if err := r.db.WithContext(ctx).Preload("Media").Preload("Language").Where("media_id = ? AND language_id = ?", mediaID, languageID).First(&subtitle).Error; err != nil {
		return nil, err
	}
	return &subtitle, nil
}

func (r *subtitleRepository) Update(ctx context.Context, subtitle *model.Subtitle) error {
	return r.db.WithContext(ctx).Save(subtitle).Error
}

func (r *subtitleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Subtitle{}, "id = ?", id).Error
}

func (r *subtitleRepository) List(ctx context.Context, filter *model.ListSubtitlesRequest) ([]*model.Subtitle, int64, error) {
	var subtitles []*model.Subtitle
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

	query := r.db.WithContext(ctx).Model(&model.Subtitle{}).Preload("Media").Preload("Language")

	if filter.MediaID != "" {
		mediaID, err := uuid.Parse(filter.MediaID)
		if err == nil {
			query = query.Where("media_id = ?", mediaID)
		}
	}

	if filter.LanguageID != "" {
		languageID, err := uuid.Parse(filter.LanguageID)
		if err == nil {
			query = query.Where("language_id = ?", languageID)
		}
	}

	if filter.Search != "" {
		search := fmt.Sprintf("%%%s%%", filter.Search)
		query = query.Where("content ILIKE ?", search)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset(offset).Find(&subtitles).Error; err != nil {
		return nil, 0, err
	}

	return subtitles, total, nil
}
