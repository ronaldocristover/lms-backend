package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/lms/internal/model"
	"gorm.io/gorm"
)

type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Session, error)
	Update(ctx context.Context, session *model.Session) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *model.ListSessionsRequest) ([]*model.Session, int64, error)
}

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *sessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Session, error) {
	var session model.Session
	if err := r.db.WithContext(ctx).Preload("Series").Where("id = ?", id).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) Update(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Save(session).Error
}

func (r *sessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Session{}, "id = ?", id).Error
}

func (r *sessionRepository) List(ctx context.Context, filter *model.ListSessionsRequest) ([]*model.Session, int64, error) {
	var sessions []*model.Session
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

	query := r.db.WithContext(ctx).Model(&model.Session{}).Preload("Series")

	if filter.Search != "" {
		search := fmt.Sprintf("%%%s%%", filter.Search)
		query = query.Where("title ILIKE ? OR description ILIKE ?", search, search)
	}

	if filter.SeriesID != uuid.Nil {
		query = query.Where("series_id = ?", filter.SeriesID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("\"order\" ASC, created_at DESC").Limit(pageSize).Offset(offset).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}
