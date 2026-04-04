package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/yourusername/lms/internal/model"
)

func setupTestDB(t *testing.T) *gorm.DB {
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
	require.NoError(t, db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			name TEXT,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role_id TEXT NOT NULL,
			organization_id TEXT,
			created_at DATETIME,
			FOREIGN KEY (role_id) REFERENCES roles(id)
		)
	`).Error)
	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	roleID := uuid.New()
	user := &model.User{
		Email:        "test@example.com",
		PasswordHash: "hashed",
		Name:         "Test User",
		RoleID:       roleID,
	}

	err := repo.Create(context.Background(), user)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	roleID := uuid.New()
	user := &model.User{
		Email:        "test@example.com",
		PasswordHash: "hashed",
		Name:         "Test User",
		RoleID:       roleID,
	}
	require.NoError(t, repo.Create(context.Background(), user))

	found, err := repo.GetByID(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.Email, found.Email)
	assert.Equal(t, user.Name, found.Name)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	roleID := uuid.New()
	user := &model.User{
		Email:        "test@example.com",
		PasswordHash: "hashed",
		Name:         "Test User",
		RoleID:       roleID,
	}
	require.NoError(t, repo.Create(context.Background(), user))

	found, err := repo.GetByEmail(context.Background(), "test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.Email, found.Email)
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.GetByEmail(context.Background(), "nonexistent@example.com")
	assert.Error(t, err)
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	roleID := uuid.New()
	user := &model.User{
		Email:        "test@example.com",
		PasswordHash: "hashed",
		Name:         "Test User",
		RoleID:       roleID,
	}
	require.NoError(t, repo.Create(context.Background(), user))

	user.Name = "Updated Name"
	err := repo.Update(context.Background(), user)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", found.Name)
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	roleID := uuid.New()
	user := &model.User{
		Email:        "test@example.com",
		PasswordHash: "hashed",
		Name:         "Test User",
		RoleID:       roleID,
	}
	require.NoError(t, repo.Create(context.Background(), user))

	err := repo.Delete(context.Background(), user.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), user.ID)
	assert.Error(t, err)
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	roleID := uuid.New()
	for i := 0; i < 5; i++ {
		user := &model.User{
			Email:        fmt.Sprintf("user%d@example.com", i),
			PasswordHash: "hashed",
			Name:         fmt.Sprintf("User %d", i),
			RoleID:       roleID,
		}
		require.NoError(t, repo.Create(context.Background(), user))
	}

	users, total, err := repo.List(context.Background(), &model.ListUsersRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Equal(t, 5, len(users))
}

func TestUserRepository_List_Pagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	roleID := uuid.New()
	for i := 0; i < 5; i++ {
		user := &model.User{
			Email:        fmt.Sprintf("user%d@example.com", i),
			PasswordHash: "hashed",
			Name:         fmt.Sprintf("User %d", i),
			RoleID:       roleID,
		}
		require.NoError(t, repo.Create(context.Background(), user))
	}

	users, total, err := repo.List(context.Background(), &model.ListUsersRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Equal(t, 2, len(users))
}

func TestUserRepository_List_DefaultPagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	users, _, err := repo.List(context.Background(), &model.ListUsersRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, users)
}

func TestUserRepository_List_WithRoleFilter(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	role1 := uuid.New()
	role2 := uuid.New()

	user1 := &model.User{Email: "user1@example.com", PasswordHash: "hashed", Name: "User1", RoleID: role1}
	user2 := &model.User{Email: "user2@example.com", PasswordHash: "hashed", Name: "User2", RoleID: role2}
	require.NoError(t, repo.Create(context.Background(), user1))
	require.NoError(t, repo.Create(context.Background(), user2))

	users, total, err := repo.List(context.Background(), &model.ListUsersRequest{
		Page:     1,
		PageSize: 20,
		RoleID:   role1.String(),
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, 1, len(users))
	assert.Equal(t, user1.ID, users[0].ID)
}

func TestUserRepository_List_WithOrgFilter(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	roleID := uuid.New()
	orgID := uuid.New()
	otherOrgID := uuid.New()

	user1 := &model.User{Email: "user1@example.com", PasswordHash: "hashed", Name: "User1", RoleID: roleID, OrganizationID: &orgID}
	user2 := &model.User{Email: "user2@example.com", PasswordHash: "hashed", Name: "User2", RoleID: roleID, OrganizationID: &otherOrgID}
	require.NoError(t, repo.Create(context.Background(), user1))
	require.NoError(t, repo.Create(context.Background(), user2))

	users, total, err := repo.List(context.Background(), &model.ListUsersRequest{
		Page:           1,
		PageSize:       20,
		OrganizationID: orgID.String(),
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, 1, len(users))
}

func TestUserRepository_List_WithSearch(t *testing.T) {
	t.Skip("ILIKE is PostgreSQL-specific; this test requires PostgreSQL")
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	roleID := uuid.New()
	user1 := &model.User{Email: "john@example.com", PasswordHash: "hashed", Name: "John Doe", RoleID: roleID}
	user2 := &model.User{Email: "jane@example.com", PasswordHash: "hashed", Name: "Jane Smith", RoleID: roleID}
	require.NoError(t, repo.Create(context.Background(), user1))
	require.NoError(t, repo.Create(context.Background(), user2))

	users, total, err := repo.List(context.Background(), &model.ListUsersRequest{
		Page:     1,
		PageSize: 20,
		Search:   "john",
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))
	assert.GreaterOrEqual(t, len(users), 1)
}

func TestNewUserRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	assert.NotNil(t, repo)
}
