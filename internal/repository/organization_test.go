package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/yourusername/lms/internal/model"
)

func setupOrgTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			name TEXT,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role_id TEXT NOT NULL,
			organization_id TEXT,
			created_at DATETIME
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE organizations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			owner_id TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE organization_users (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'member',
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error)
	return db
}

func TestOrganizationRepository_Create(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationRepository(db)

	org := &model.Organization{
		Name:    "Test Org",
		OwnerID: uuid.New(),
	}

	err := repo.Create(context.Background(), org)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, org.ID)
}

func TestOrganizationRepository_GetByID(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationRepository(db)

	ownerID := uuid.New()
	org := &model.Organization{
		Name:    "Test Org",
		OwnerID: ownerID,
	}
	require.NoError(t, repo.Create(context.Background(), org))

	found, err := repo.GetByID(context.Background(), org.ID)
	assert.NoError(t, err)
	assert.Equal(t, org.ID, found.ID)
	assert.Equal(t, org.Name, found.Name)
	assert.Equal(t, ownerID, found.OwnerID)
}

func TestOrganizationRepository_GetByID_NotFound(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestOrganizationRepository_GetByIDWithOwner(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationRepository(db)

	ownerID := uuid.New()
	org := &model.Organization{
		Name:    "Test Org",
		OwnerID: ownerID,
	}
	require.NoError(t, repo.Create(context.Background(), org))

	found, err := repo.GetByIDWithOwner(context.Background(), org.ID)
	assert.NoError(t, err)
	assert.Equal(t, org.ID, found.ID)
}

func TestOrganizationRepository_GetByIDWithOwner_NotFound(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationRepository(db)

	_, err := repo.GetByIDWithOwner(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestOrganizationRepository_Update(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationRepository(db)

	org := &model.Organization{
		Name:    "Test Org",
		OwnerID: uuid.New(),
	}
	require.NoError(t, repo.Create(context.Background(), org))

	org.Name = "Updated Org"
	err := repo.Update(context.Background(), org)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), org.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Org", found.Name)
}

func TestOrganizationRepository_Delete(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationRepository(db)

	org := &model.Organization{
		Name:    "Test Org",
		OwnerID: uuid.New(),
	}
	require.NoError(t, repo.Create(context.Background(), org))

	err := repo.Delete(context.Background(), org.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), org.ID)
	assert.Error(t, err)
}

func TestOrganizationRepository_List(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationRepository(db)

	ownerID := uuid.New()
	for i := 0; i < 3; i++ {
		org := &model.Organization{
			Name:    "Org %d",
			OwnerID: ownerID,
		}
		require.NoError(t, repo.Create(context.Background(), org))
	}

	orgs, total, err := repo.List(context.Background(), &model.ListOrganizationsRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Equal(t, 3, len(orgs))
}

func TestOrganizationRepository_List_Pagination(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationRepository(db)

	ownerID := uuid.New()
	for i := 0; i < 5; i++ {
		org := &model.Organization{
			Name:    "Org %d",
			OwnerID: ownerID,
		}
		require.NoError(t, repo.Create(context.Background(), org))
	}

	orgs, total, err := repo.List(context.Background(), &model.ListOrganizationsRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Equal(t, 2, len(orgs))
}

func TestOrganizationRepository_List_DefaultPagination(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationRepository(db)

	orgs, _, err := repo.List(context.Background(), &model.ListOrganizationsRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, orgs)
}

func TestOrganizationRepository_List_WithOwnerFilter(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationRepository(db)

	owner1 := uuid.New()
	owner2 := uuid.New()

	org1 := &model.Organization{Name: "Org1", OwnerID: owner1}
	org2 := &model.Organization{Name: "Org2", OwnerID: owner2}
	require.NoError(t, repo.Create(context.Background(), org1))
	require.NoError(t, repo.Create(context.Background(), org2))

	orgs, total, err := repo.List(context.Background(), &model.ListOrganizationsRequest{
		Page:     1,
		PageSize: 20,
		OwnerID:  owner1.String(),
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, 1, len(orgs))
}

func TestOrganizationRepository_List_WithSearch(t *testing.T) {
	t.Skip("ILIKE is PostgreSQL-specific; this test requires PostgreSQL")
	db := setupOrgTestDB(t)
	repo := NewOrganizationRepository(db)

	ownerID := uuid.New()
	org1 := &model.Organization{Name: "Acme Corp", OwnerID: ownerID}
	org2 := &model.Organization{Name: "Beta Inc", OwnerID: ownerID}
	require.NoError(t, repo.Create(context.Background(), org1))
	require.NoError(t, repo.Create(context.Background(), org2))

	orgs, total, err := repo.List(context.Background(), &model.ListOrganizationsRequest{
		Page:     1,
		PageSize: 20,
		Search:   "acme",
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))
	assert.GreaterOrEqual(t, len(orgs), 1)
}

func TestOrganizationRepository_ExistsByName(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationRepository(db)

	org := &model.Organization{
		Name:    "Existing Org",
		OwnerID: uuid.New(),
	}
	require.NoError(t, repo.Create(context.Background(), org))

	exists, err := repo.ExistsByName(context.Background(), "Existing Org")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.ExistsByName(context.Background(), "Nonexistent")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestOrganizationUserRepository_Create(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationUserRepository(db)

	orgID := uuid.New()
	userID := uuid.New()
	orgUser := &model.OrganizationUser{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           model.OrgRoleMember,
	}

	err := repo.Create(context.Background(), orgUser)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, orgUser.ID)
}

func TestOrganizationUserRepository_GetByID(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationUserRepository(db)

	orgID := uuid.New()
	userID := uuid.New()
	orgUser := &model.OrganizationUser{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           model.OrgRoleMember,
	}
	require.NoError(t, repo.Create(context.Background(), orgUser))

	found, err := repo.GetByID(context.Background(), orgUser.ID)
	assert.NoError(t, err)
	assert.Equal(t, orgUser.ID, found.ID)
}

func TestOrganizationUserRepository_GetByID_NotFound(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationUserRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestOrganizationUserRepository_GetByOrgAndUser(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationUserRepository(db)

	orgID := uuid.New()
	userID := uuid.New()
	orgUser := &model.OrganizationUser{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           model.OrgRoleAdmin,
	}
	require.NoError(t, repo.Create(context.Background(), orgUser))

	found, err := repo.GetByOrgAndUser(context.Background(), orgID, userID)
	assert.NoError(t, err)
	assert.Equal(t, orgUser.ID, found.ID)
	assert.Equal(t, model.OrgRoleAdmin, found.Role)
}

func TestOrganizationUserRepository_GetByOrgAndUser_NotFound(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationUserRepository(db)

	_, err := repo.GetByOrgAndUser(context.Background(), uuid.New(), uuid.New())
	assert.Error(t, err)
}

func TestOrganizationUserRepository_Update(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationUserRepository(db)

	orgID := uuid.New()
	userID := uuid.New()
	orgUser := &model.OrganizationUser{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           model.OrgRoleMember,
	}
	require.NoError(t, repo.Create(context.Background(), orgUser))

	orgUser.Role = model.OrgRoleAdmin
	err := repo.Update(context.Background(), orgUser)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), orgUser.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.OrgRoleAdmin, found.Role)
}

func TestOrganizationUserRepository_Delete(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationUserRepository(db)

	orgID := uuid.New()
	userID := uuid.New()
	orgUser := &model.OrganizationUser{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           model.OrgRoleMember,
	}
	require.NoError(t, repo.Create(context.Background(), orgUser))

	err := repo.Delete(context.Background(), orgUser.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), orgUser.ID)
	assert.Error(t, err)
}

func TestOrganizationUserRepository_ListByOrganization(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationUserRepository(db)

	orgID := uuid.New()
	for i := 0; i < 3; i++ {
		orgUser := &model.OrganizationUser{
			OrganizationID: orgID,
			UserID:         uuid.New(),
			Role:           model.OrgRoleMember,
		}
		require.NoError(t, repo.Create(context.Background(), orgUser))
	}

	users, total, err := repo.ListByOrganization(context.Background(), orgID, &model.ListOrgUsersRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Equal(t, 3, len(users))
}

func TestOrganizationUserRepository_ListByOrganization_Pagination(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationUserRepository(db)

	orgID := uuid.New()
	for i := 0; i < 5; i++ {
		orgUser := &model.OrganizationUser{
			OrganizationID: orgID,
			UserID:         uuid.New(),
			Role:           model.OrgRoleMember,
		}
		require.NoError(t, repo.Create(context.Background(), orgUser))
	}

	users, total, err := repo.ListByOrganization(context.Background(), orgID, &model.ListOrgUsersRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Equal(t, 2, len(users))
}

func TestOrganizationUserRepository_ListByOrganization_DefaultPagination(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationUserRepository(db)

	users, _, err := repo.ListByOrganization(context.Background(), uuid.New(), &model.ListOrgUsersRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, users)
}

func TestOrganizationUserRepository_ListByOrganization_WithRoleFilter(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationUserRepository(db)

	orgID := uuid.New()
	orgUser1 := &model.OrganizationUser{OrganizationID: orgID, UserID: uuid.New(), Role: model.OrgRoleAdmin}
	orgUser2 := &model.OrganizationUser{OrganizationID: orgID, UserID: uuid.New(), Role: model.OrgRoleMember}
	require.NoError(t, repo.Create(context.Background(), orgUser1))
	require.NoError(t, repo.Create(context.Background(), orgUser2))

	users, total, err := repo.ListByOrganization(context.Background(), orgID, &model.ListOrgUsersRequest{
		Page:     1,
		PageSize: 20,
		Role:     model.OrgRoleAdmin,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, 1, len(users))
}

func TestOrganizationUserRepository_ExistsByOrgAndUser(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationUserRepository(db)

	orgID := uuid.New()
	userID := uuid.New()
	orgUser := &model.OrganizationUser{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           model.OrgRoleMember,
	}
	require.NoError(t, repo.Create(context.Background(), orgUser))

	exists, err := repo.ExistsByOrgAndUser(context.Background(), orgID, userID)
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.ExistsByOrgAndUser(context.Background(), orgID, uuid.New())
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestOrganizationUserRepository_DeleteByOrganization(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationUserRepository(db)

	orgID := uuid.New()
	for i := 0; i < 3; i++ {
		orgUser := &model.OrganizationUser{
			OrganizationID: orgID,
			UserID:         uuid.New(),
			Role:           model.OrgRoleMember,
		}
		require.NoError(t, repo.Create(context.Background(), orgUser))
	}

	err := repo.DeleteByOrganization(context.Background(), orgID)
	assert.NoError(t, err)

	users, total, err := repo.ListByOrganization(context.Background(), orgID, &model.ListOrgUsersRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Equal(t, 0, len(users))
}

func TestNewOrganizationRepository(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationRepository(db)
	assert.NotNil(t, repo)
}

func TestNewOrganizationUserRepository(t *testing.T) {
	db := setupOrgTestDB(t)
	repo := NewOrganizationUserRepository(db)
	assert.NotNil(t, repo)
}
