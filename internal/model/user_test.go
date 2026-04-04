package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUser_BeforeCreate_WithNilID(t *testing.T) {
	user := User{
		Email:        "test@example.com",
		PasswordHash: "hashed",
		Name:         "Test User",
	}

	assert.Equal(t, uuid.Nil, user.ID)

	err := user.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
}

func TestUser_BeforeCreate_WithExistingID(t *testing.T) {
	existingID := uuid.New()
	user := User{
		ID:           existingID,
		Email:        "test@example.com",
		PasswordHash: "hashed",
		Name:         "Test User",
	}

	err := user.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.Equal(t, existingID, user.ID)
}

func TestUserConstants(t *testing.T) {
	assert.Equal(t, "active", UserStatusActive)
	assert.Equal(t, "inactive", UserStatusInactive)
	assert.Equal(t, "suspended", UserStatusSuspended)

	assert.Equal(t, "admin", UserRoleAdmin)
	assert.Equal(t, "user", UserRoleUser)
	assert.Equal(t, "tutor", UserRoleTutor)
}

func TestUser_JSONTags(t *testing.T) {
	user := User{
		Email:        "test@example.com",
		PasswordHash: "secret",
		Name:         "Test",
		Role:         "admin",
		Status:       "active",
	}

	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "secret", user.PasswordHash)
	assert.Equal(t, "Test", user.Name)
	assert.Equal(t, "admin", user.Role)
	assert.Equal(t, "active", user.Status)
}
