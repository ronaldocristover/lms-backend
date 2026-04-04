package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestOrganization_BeforeCreate_WithNilID(t *testing.T) {
	org := Organization{
		Name:    "Test Org",
		OwnerID: uuid.New(),
	}

	assert.Equal(t, uuid.Nil, org.ID)

	err := org.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, org.ID)
}

func TestOrganization_BeforeCreate_WithExistingID(t *testing.T) {
	existingID := uuid.New()
	org := Organization{
		ID:      existingID,
		Name:    "Test Org",
		OwnerID: uuid.New(),
	}

	err := org.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.Equal(t, existingID, org.ID)
}

func TestOrganizationUser_BeforeCreate_WithNilID(t *testing.T) {
	orgUser := OrganizationUser{
		OrganizationID: uuid.New(),
		UserID:         uuid.New(),
		Role:           OrgRoleMember,
	}

	assert.Equal(t, uuid.Nil, orgUser.ID)

	err := orgUser.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, orgUser.ID)
}

func TestOrganizationUser_BeforeCreate_WithExistingID(t *testing.T) {
	existingID := uuid.New()
	orgUser := OrganizationUser{
		ID:             existingID,
		OrganizationID: uuid.New(),
		UserID:         uuid.New(),
		Role:           OrgRoleAdmin,
	}

	err := orgUser.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.Equal(t, existingID, orgUser.ID)
}

func TestOrgRoleConstants(t *testing.T) {
	assert.Equal(t, "org_admin", OrgRoleAdmin)
	assert.Equal(t, "member", OrgRoleMember)
}

func TestOrganization_JSONFields(t *testing.T) {
	ownerID := uuid.New()
	org := Organization{
		Name:    "Test Org",
		OwnerID: ownerID,
	}

	assert.Equal(t, "Test Org", org.Name)
	assert.Equal(t, ownerID, org.OwnerID)
}

func TestOrganizationUser_JSONFields(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	orgUser := OrganizationUser{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           OrgRoleMember,
	}

	assert.Equal(t, orgID, orgUser.OrganizationID)
	assert.Equal(t, userID, orgUser.UserID)
	assert.Equal(t, OrgRoleMember, orgUser.Role)
}

func TestCreateOrganizationRequest(t *testing.T) {
	req := CreateOrganizationRequest{
		Name:    "New Org",
		OwnerID: uuid.New().String(),
	}

	assert.Equal(t, "New Org", req.Name)
	assert.NotEmpty(t, req.OwnerID)
}

func TestUpdateOrganizationRequest(t *testing.T) {
	req := UpdateOrganizationRequest{
		Name: "Updated Org",
	}

	assert.Equal(t, "Updated Org", req.Name)
}

func TestAddOrgUserRequest(t *testing.T) {
	req := AddOrgUserRequest{
		UserID: uuid.New().String(),
		Role:   OrgRoleAdmin,
	}

	assert.NotEmpty(t, req.UserID)
	assert.Equal(t, OrgRoleAdmin, req.Role)
}

func TestUpdateOrgUserRoleRequest(t *testing.T) {
	req := UpdateOrgUserRoleRequest{
		Role: OrgRoleMember,
	}

	assert.Equal(t, OrgRoleMember, req.Role)
}

func TestListOrganizationsRequest(t *testing.T) {
	req := ListOrganizationsRequest{
		Page:     1,
		PageSize: 20,
		Search:   "test",
		OwnerID:  uuid.New().String(),
		UserID:   uuid.New().String(),
	}

	assert.Equal(t, 1, req.Page)
	assert.Equal(t, 20, req.PageSize)
	assert.Equal(t, "test", req.Search)
}

func TestListOrgUsersRequest(t *testing.T) {
	req := ListOrgUsersRequest{
		Page:     2,
		PageSize: 50,
		Role:     OrgRoleAdmin,
		Search:   "john",
	}

	assert.Equal(t, 2, req.Page)
	assert.Equal(t, 50, req.PageSize)
	assert.Equal(t, OrgRoleAdmin, req.Role)
	assert.Equal(t, "john", req.Search)
}
