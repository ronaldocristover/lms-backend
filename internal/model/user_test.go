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

func TestRoleConstants(t *testing.T) {
	assert.Equal(t, "admin", RoleAdmin)
	assert.Equal(t, "student", RoleStudent)
	assert.Equal(t, "tutor", RoleTutor)
	assert.Equal(t, "org_admin", RoleOrgAdmin)
}

func TestUser_JSONTags(t *testing.T) {
	roleID := uuid.New()
	orgID := uuid.New()
	user := User{
		Email:          "test@example.com",
		PasswordHash:   "secret",
		Name:           "Test",
		RoleID:         roleID,
		OrganizationID: &orgID,
	}

	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "secret", user.PasswordHash)
	assert.Equal(t, "Test", user.Name)
	assert.Equal(t, roleID, user.RoleID)
	assert.Equal(t, &orgID, user.OrganizationID)
}

func TestRole_BeforeCreate_WithNilID(t *testing.T) {
	role := Role{
		Name: RoleAdmin,
	}

	assert.Equal(t, uuid.Nil, role.ID)

	err := role.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, role.ID)
}

func TestRole_BeforeCreate_WithExistingID(t *testing.T) {
	existingID := uuid.New()
	role := Role{
		ID:   existingID,
		Name: RoleStudent,
	}

	err := role.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.Equal(t, existingID, role.ID)
}
