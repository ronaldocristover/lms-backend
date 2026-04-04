package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null;size:50" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

const (
	RoleAdmin    = "admin"
	RoleStudent  = "student"
	RoleTutor    = "tutor"
	RoleOrgAdmin = "org_admin"
)

type CreateRoleRequest struct {
	Name string `json:"name" binding:"required,oneof=admin student tutor org_admin"`
}

type UpdateRoleRequest struct {
	Name string `json:"name" binding:"required,oneof=admin student tutor org_admin"`
}

type ListRolesRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Search   string `form:"search" binding:"omitempty,max=255"`
}
