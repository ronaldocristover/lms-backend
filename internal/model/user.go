package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash string    `gorm:"not null;size:255" json:"-"`
	Name         string    `gorm:"size:255" json:"name"`
	Role         string    `gorm:"size:50;default:'user'" json:"role"`
	Avatar       string    `gorm:"size:500" json:"avatar"`
	Status       string    `gorm:"size:20;default:'active'" json:"status"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

const (
	UserStatusActive    = "active"
	UserStatusInactive  = "inactive"
	UserStatusSuspended = "suspended"

	UserRoleAdmin  = "admin"
	UserRoleUser   = "user"
	UserRoleTutor  = "tutor"
)

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type UpdateUserRequest struct {
	Name   string `json:"name" binding:"omitempty,min=1,max=255"`
	Role   string `json:"role" binding:"omitempty,oneof=admin user tutor"`
	Avatar string `json:"avatar" binding:"omitempty,url,max=500"`
	Status string `json:"status" binding:"omitempty,oneof=active inactive suspended"`
}

type ListUsersRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Role     string `form:"role" binding:"omitempty,oneof=admin user tutor"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive suspended"`
	Search   string `form:"search" binding:"omitempty,max=255"`
}
