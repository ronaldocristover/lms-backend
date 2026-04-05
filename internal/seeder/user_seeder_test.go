package seeder

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

func setupSeederTestDB(t *testing.T) *gorm.DB {
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
			created_at DATETIME
		)
	`).Error)
	return db
}

func TestNewUserSeeder(t *testing.T) {
	db := setupSeederTestDB(t)
	seeder := NewUserSeeder(db)

	assert.NotNil(t, seeder)
	assert.Equal(t, db, seeder.db)
}

func TestUserSeeder_Seed(t *testing.T) {
	db := setupSeederTestDB(t)
	seeder := NewUserSeeder(db)

	err := seeder.Seed(context.Background())
	assert.NoError(t, err)

	var count int64
	db.Model(&model.User{}).Count(&count)
	assert.Equal(t, int64(4), count)

	var roleCount int64
	db.Model(&model.Role{}).Count(&roleCount)
	assert.Equal(t, int64(4), roleCount)
}

func TestUserSeeder_Seed_Idempotent(t *testing.T) {
	db := setupSeederTestDB(t)
	seeder := NewUserSeeder(db)

	err := seeder.Seed(context.Background())
	assert.NoError(t, err)

	err = seeder.Seed(context.Background())
	assert.NoError(t, err)

	var count int64
	db.Model(&model.User{}).Count(&count)
	assert.Equal(t, int64(4), count)
}

func TestUserSeeder_Truncate(t *testing.T) {
	db := setupSeederTestDB(t)
	seeder := NewUserSeeder(db)

	err := seeder.Seed(context.Background())
	require.NoError(t, err)

	err = seeder.Truncate(context.Background())
	assert.NoError(t, err)

	var userCount int64
	db.Model(&model.User{}).Count(&userCount)
	assert.Equal(t, int64(0), userCount)

	var roleCount int64
	db.Model(&model.Role{}).Count(&roleCount)
	assert.Equal(t, int64(0), roleCount)
}

func TestUserSeeder_GetOrCreateRole_Existing(t *testing.T) {
	db := setupSeederTestDB(t)
	seeder := NewUserSeeder(db)

	role := &model.Role{Name: model.RoleAdmin}
	require.NoError(t, db.Create(role).Error)

	found, err := seeder.getOrCreateRole(context.Background(), model.RoleAdmin)
	assert.NoError(t, err)
	assert.Equal(t, model.RoleAdmin, found.Name)
}

func TestUserSeeder_GetOrCreateRole_New(t *testing.T) {
	db := setupSeederTestDB(t)
	seeder := NewUserSeeder(db)

	found, err := seeder.getOrCreateRole(context.Background(), model.RoleStudent)
	assert.NoError(t, err)
	assert.Equal(t, model.RoleStudent, found.Name)
	assert.NotEqual(t, uuid.Nil, found.ID)
}
