package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Organization struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"not null;size:255" json:"name"`
	OwnerID   uuid.UUID      `gorm:"type:uuid;not null" json:"owner_id"`
	Owner     *User          `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Users     []OrganizationUser `gorm:"foreignKey:OrganizationID" json:"users,omitempty"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (o *Organization) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

type OrganizationUser struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null" json:"organization_id"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	UserID         uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User           *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role           string    `gorm:"size:20;not null;default:'member'" json:"role"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ou *OrganizationUser) BeforeCreate(tx *gorm.DB) error {
	if ou.ID == uuid.Nil {
		ou.ID = uuid.New()
	}
	return nil
}

const (
	OrgRoleAdmin  = "org_admin"
	OrgRoleMember = "member"
)

type CreateOrganizationRequest struct {
	Name    string `json:"name" binding:"required,min=1,max=255"`
	OwnerID string `json:"owner_id" binding:"required,uuid"`
}

type UpdateOrganizationRequest struct {
	Name string `json:"name" binding:"omitempty,min=1,max=255"`
}

type AddOrgUserRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
	Role   string `json:"role" binding:"required,oneof=org_admin member"`
}

type UpdateOrgUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=org_admin member"`
}

type ListOrganizationsRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Search   string `form:"search" binding:"omitempty,max=255"`
	OwnerID  string `form:"owner_id" binding:"omitempty,uuid"`
	UserID   string `form:"user_id" binding:"omitempty,uuid"`
}

type ListOrgUsersRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Role     string `form:"role" binding:"omitempty,oneof=org_admin member"`
	Search   string `form:"search" binding:"omitempty,max=255"`
}
