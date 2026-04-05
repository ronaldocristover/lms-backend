package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ronaldocristover/lms-backend/internal/model"
	"gorm.io/gorm"
)

type OrganizationRepository interface {
	Create(ctx context.Context, org *model.Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error)
	GetByIDWithOwner(ctx context.Context, id uuid.UUID) (*model.Organization, error)
	Update(ctx context.Context, org *model.Organization) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *model.ListOrganizationsRequest) ([]*model.Organization, int64, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
}

type organizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) OrganizationRepository {
	return &organizationRepository{db: db}
}

func (r *organizationRepository) Create(ctx context.Context, org *model.Organization) error {
	return r.db.WithContext(ctx).Create(org).Error
}

func (r *organizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	var org model.Organization
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&org).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *organizationRepository) GetByIDWithOwner(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	var org model.Organization
	if err := r.db.WithContext(ctx).Preload("Owner").Where("id = ?", id).First(&org).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *organizationRepository) Update(ctx context.Context, org *model.Organization) error {
	return r.db.WithContext(ctx).Save(org).Error
}

func (r *organizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("organization_id = ?", id).Delete(&model.OrganizationUser{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Organization{}, "id = ?", id).Error
	})
}

func (r *organizationRepository) List(ctx context.Context, filter *model.ListOrganizationsRequest) ([]*model.Organization, int64, error) {
	var orgs []*model.Organization
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

	query := r.db.WithContext(ctx).Model(&model.Organization{})

	if filter.OwnerID != "" {
		ownerID, err := uuid.Parse(filter.OwnerID)
		if err == nil {
			query = query.Where("owner_id = ?", ownerID)
		}
	}
	if filter.Search != "" {
		search := fmt.Sprintf("%%%s%%", filter.Search)
		query = query.Where("name ILIKE ?", search)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Owner").Limit(pageSize).Offset(offset).Find(&orgs).Error; err != nil {
		return nil, 0, err
	}

	return orgs, total, nil
}

func (r *organizationRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Organization{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

type OrganizationUserRepository interface {
	Create(ctx context.Context, orgUser *model.OrganizationUser) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.OrganizationUser, error)
	GetByOrgAndUser(ctx context.Context, orgID, userID uuid.UUID) (*model.OrganizationUser, error)
	Update(ctx context.Context, orgUser *model.OrganizationUser) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByOrganization(ctx context.Context, orgID uuid.UUID, filter *model.ListOrgUsersRequest) ([]*model.OrganizationUser, int64, error)
	ExistsByOrgAndUser(ctx context.Context, orgID, userID uuid.UUID) (bool, error)
	DeleteByOrganization(ctx context.Context, orgID uuid.UUID) error
}

type organizationUserRepository struct {
	db *gorm.DB
}

func NewOrganizationUserRepository(db *gorm.DB) OrganizationUserRepository {
	return &organizationUserRepository{db: db}
}

func (r *organizationUserRepository) Create(ctx context.Context, orgUser *model.OrganizationUser) error {
	return r.db.WithContext(ctx).Create(orgUser).Error
}

func (r *organizationUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.OrganizationUser, error) {
	var orgUser model.OrganizationUser
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&orgUser).Error; err != nil {
		return nil, err
	}
	return &orgUser, nil
}

func (r *organizationUserRepository) GetByOrgAndUser(ctx context.Context, orgID, userID uuid.UUID) (*model.OrganizationUser, error) {
	var orgUser model.OrganizationUser
	if err := r.db.WithContext(ctx).Where("organization_id = ? AND user_id = ?", orgID, userID).First(&orgUser).Error; err != nil {
		return nil, err
	}
	return &orgUser, nil
}

func (r *organizationUserRepository) Update(ctx context.Context, orgUser *model.OrganizationUser) error {
	return r.db.WithContext(ctx).Save(orgUser).Error
}

func (r *organizationUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.OrganizationUser{}, "id = ?", id).Error
}

func (r *organizationUserRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID, filter *model.ListOrgUsersRequest) ([]*model.OrganizationUser, int64, error) {
	var orgUsers []*model.OrganizationUser
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

	query := r.db.WithContext(ctx).Model(&model.OrganizationUser{}).Where("organization_id = ?", orgID)

	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}
	if filter.Search != "" {
		search := fmt.Sprintf("%%%s%%", filter.Search)
		query = query.Joins("JOIN users ON users.id = organization_users.user_id").
			Where("users.name ILIKE ? OR users.email ILIKE ?", search, search)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("User").Limit(pageSize).Offset(offset).Find(&orgUsers).Error; err != nil {
		return nil, 0, err
	}

	return orgUsers, total, nil
}

func (r *organizationUserRepository) ExistsByOrgAndUser(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.OrganizationUser{}).Where("organization_id = ? AND user_id = ?", orgID, userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *organizationUserRepository) DeleteByOrganization(ctx context.Context, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("organization_id = ?", orgID).Delete(&model.OrganizationUser{}).Error
}
