package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name           string     `gorm:"size:255" json:"name"`
	Email          string     `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash   string     `gorm:"not null;size:255" json:"-"`
	RoleID         uuid.UUID  `gorm:"type:uuid;not null" json:"role_id"`
	Role           *Role      `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	OrganizationID *uuid.UUID `gorm:"type:uuid" json:"organization_id,omitempty"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type RegisterRequest struct {
	Email          string     `json:"email" binding:"required,email"`
	Password       string     `json:"password" binding:"required,min=8"`
	Name           string     `json:"name" binding:"required,min=1,max=255"`
	RoleID         uuid.UUID  `json:"role_id" binding:"required"`
	OrganizationID *uuid.UUID `json:"organization_id"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type CreateUserRequest struct {
	Name           string     `json:"name" binding:"required,min=1,max=255"`
	Email          string     `json:"email" binding:"required,email"`
	Password       string     `json:"password" binding:"required,min=8"`
	RoleID         uuid.UUID  `json:"role_id" binding:"required"`
	OrganizationID *uuid.UUID `json:"organization_id"`
}

type UpdateUserRequest struct {
	Name           string     `json:"name" binding:"omitempty,min=1,max=255"`
	Email          string     `json:"email" binding:"omitempty,email"`
	RoleID         uuid.UUID  `json:"role_id"`
	OrganizationID *uuid.UUID `json:"organization_id"`
}

type ListUsersRequest struct {
	Page           int    `form:"page" binding:"omitempty,min=1"`
	PageSize       int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	RoleID         string `form:"role_id" binding:"omitempty"`
	OrganizationID string `form:"organization_id" binding:"omitempty"`
	Search         string `form:"search" binding:"omitempty,max=255"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
