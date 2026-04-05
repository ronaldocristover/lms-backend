package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/lms/internal/model"
	"github.com/yourusername/lms/pkg/pagination"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *model.ListUsersRequest) ([]*model.User, int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Preload("Role").Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Preload("Role").Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, "id = ?", id).Error
}

func (r *userRepository) List(ctx context.Context, filter *model.ListUsersRequest) ([]*model.User, int64, error) {
	var users []*model.User
	query := r.db.WithContext(ctx).Model(&model.User{})

	if filter.RoleID != "" {
		if roleID, err := uuid.Parse(filter.RoleID); err == nil {
			query = query.Where("role_id = ?", roleID)
		}
	}
	if filter.OrganizationID != "" {
		if orgID, err := uuid.Parse(filter.OrganizationID); err == nil {
			query = query.Where("organization_id = ?", orgID)
		}
	}
	if filter.Search != "" {
		search := fmt.Sprintf("%%%s%%", filter.Search)
		query = query.Where("name ILIKE ? OR email ILIKE ?", search, search)
	}

	total, paginated, err := pagination.Paginate(query, filter.Page, filter.PageSize)
	if err != nil {
		return nil, 0, err
	}

	if err := paginated.Preload("Role").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
