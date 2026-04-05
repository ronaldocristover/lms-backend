package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ronaldocristover/lms-backend/internal/model"
	"github.com/ronaldocristover/lms-backend/pkg/pagination"
	"gorm.io/gorm"
)

type LanguageRepository interface {
	Create(ctx context.Context, language *model.Language) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Language, error)
	GetByCode(ctx context.Context, code string) (*model.Language, error)
	Update(ctx context.Context, language *model.Language) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *model.ListLanguagesRequest) ([]*model.Language, int64, error)
}

type languageRepository struct {
	db *gorm.DB
}

func NewLanguageRepository(db *gorm.DB) LanguageRepository {
	return &languageRepository{db: db}
}

func (r *languageRepository) Create(ctx context.Context, language *model.Language) error {
	return r.db.WithContext(ctx).Create(language).Error
}

func (r *languageRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Language, error) {
	var language model.Language
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&language).Error; err != nil {
		return nil, err
	}
	return &language, nil
}

func (r *languageRepository) GetByCode(ctx context.Context, code string) (*model.Language, error) {
	var language model.Language
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&language).Error; err != nil {
		return nil, err
	}
	return &language, nil
}

func (r *languageRepository) Update(ctx context.Context, language *model.Language) error {
	return r.db.WithContext(ctx).Save(language).Error
}

func (r *languageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Language{}, "id = ?", id).Error
}

func (r *languageRepository) List(ctx context.Context, filter *model.ListLanguagesRequest) ([]*model.Language, int64, error) {
	var languages []*model.Language
	query := r.db.WithContext(ctx).Model(&model.Language{})

	if filter.Search != "" {
		search := fmt.Sprintf("%%%s%%", filter.Search)
		query = query.Where("code ILIKE ? OR name ILIKE ?", search, search)
	}

	total, paginated, err := pagination.Paginate(query, filter.Page, filter.PageSize)
	if err != nil {
		return nil, 0, err
	}

	if err := paginated.Find(&languages).Error; err != nil {
		return nil, 0, err
	}

	return languages, total, nil
}
