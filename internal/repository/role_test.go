package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/ronaldocristover/lms-backend/internal/model"
)

func setupRoleTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE roles (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			created_at DATETIME
		)
	`).Error)
	return db
}

func TestRoleRepository_Create(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := NewRoleRepository(db)

	role := &model.Role{
		Name: model.RoleAdmin,
	}

	err := repo.Create(context.Background(), role)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, role.ID)
}

func TestRoleRepository_GetByID(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := NewRoleRepository(db)

	role := &model.Role{
		Name: model.RoleStudent,
	}
	require.NoError(t, repo.Create(context.Background(), role))

	found, err := repo.GetByID(context.Background(), role.ID)
	assert.NoError(t, err)
	assert.Equal(t, role.ID, found.ID)
	assert.Equal(t, model.RoleStudent, found.Name)
}

func TestRoleRepository_GetByID_NotFound(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := NewRoleRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestRoleRepository_GetByName(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := NewRoleRepository(db)

	role := &model.Role{
		Name: model.RoleTutor,
	}
	require.NoError(t, repo.Create(context.Background(), role))

	found, err := repo.GetByName(context.Background(), model.RoleTutor)
	assert.NoError(t, err)
	assert.Equal(t, role.ID, found.ID)
	assert.Equal(t, model.RoleTutor, found.Name)
}

func TestRoleRepository_GetByName_NotFound(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := NewRoleRepository(db)

	_, err := repo.GetByName(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestRoleRepository_Update(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := NewRoleRepository(db)

	role := &model.Role{
		Name: model.RoleStudent,
	}
	require.NoError(t, repo.Create(context.Background(), role))

	role.Name = model.RoleTutor
	err := repo.Update(context.Background(), role)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), role.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.RoleTutor, found.Name)
}

func TestRoleRepository_Delete(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := NewRoleRepository(db)

	role := &model.Role{
		Name: model.RoleAdmin,
	}
	require.NoError(t, repo.Create(context.Background(), role))

	err := repo.Delete(context.Background(), role.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), role.ID)
	assert.Error(t, err)
}

func TestRoleRepository_List(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := NewRoleRepository(db)

	roles := []string{model.RoleAdmin, model.RoleStudent, model.RoleTutor, model.RoleOrgAdmin}
	for _, name := range roles {
		role := &model.Role{Name: name}
		require.NoError(t, repo.Create(context.Background(), role))
	}

	result, total, err := repo.List(context.Background(), &model.ListRolesRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(4), total)
	assert.Equal(t, 4, len(result))
}

func TestRoleRepository_List_Pagination(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := NewRoleRepository(db)

	roles := []string{model.RoleAdmin, model.RoleStudent, model.RoleTutor, model.RoleOrgAdmin}
	for _, name := range roles {
		role := &model.Role{Name: name}
		require.NoError(t, repo.Create(context.Background(), role))
	}

	result, total, err := repo.List(context.Background(), &model.ListRolesRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(4), total)
	assert.Equal(t, 2, len(result))
}

func TestRoleRepository_List_DefaultPagination(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := NewRoleRepository(db)

	result, _, err := repo.List(context.Background(), &model.ListRolesRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestRoleRepository_List_WithSearch(t *testing.T) {
	t.Skip("ILIKE is PostgreSQL-specific; this test requires PostgreSQL")
	db := setupRoleTestDB(t)
	repo := NewRoleRepository(db)

	for _, name := range []string{model.RoleAdmin, model.RoleStudent, model.RoleTutor} {
		role := &model.Role{Name: name}
		require.NoError(t, repo.Create(context.Background(), role))
	}

	result, total, err := repo.List(context.Background(), &model.ListRolesRequest{
		Page:     1,
		PageSize: 20,
		Search:   "admin",
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))
	assert.GreaterOrEqual(t, len(result), 1)
}

func TestNewRoleRepository(t *testing.T) {
	db := setupRoleTestDB(t)
	repo := NewRoleRepository(db)
	assert.NotNil(t, repo)
}
