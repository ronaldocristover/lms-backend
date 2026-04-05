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

func setupCategoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE categories (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error)
	return db
}

func TestCategoryRepository_Create(t *testing.T) {
	db := setupCategoryTestDB(t)
	repo := NewCategoryRepository(db)

	category := &model.Category{
		Name: "Mathematics",
	}

	err := repo.Create(context.Background(), category)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, category.ID)
}

func TestCategoryRepository_GetByID(t *testing.T) {
	db := setupCategoryTestDB(t)
	repo := NewCategoryRepository(db)

	category := &model.Category{
		Name: "Science",
	}
	require.NoError(t, repo.Create(context.Background(), category))

	found, err := repo.GetByID(context.Background(), category.ID)
	assert.NoError(t, err)
	assert.Equal(t, category.ID, found.ID)
	assert.Equal(t, "Science", found.Name)
}

func TestCategoryRepository_GetByID_NotFound(t *testing.T) {
	db := setupCategoryTestDB(t)
	repo := NewCategoryRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestCategoryRepository_GetByName(t *testing.T) {
	db := setupCategoryTestDB(t)
	repo := NewCategoryRepository(db)

	category := &model.Category{
		Name: "Programming",
	}
	require.NoError(t, repo.Create(context.Background(), category))

	found, err := repo.GetByName(context.Background(), "Programming")
	assert.NoError(t, err)
	assert.Equal(t, category.ID, found.ID)
	assert.Equal(t, "Programming", found.Name)
}

func TestCategoryRepository_GetByName_NotFound(t *testing.T) {
	db := setupCategoryTestDB(t)
	repo := NewCategoryRepository(db)

	_, err := repo.GetByName(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestCategoryRepository_Update(t *testing.T) {
	db := setupCategoryTestDB(t)
	repo := NewCategoryRepository(db)

	category := &model.Category{
		Name: "Mathematics",
	}
	require.NoError(t, repo.Create(context.Background(), category))

	category.Name = "Advanced Mathematics"
	err := repo.Update(context.Background(), category)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), category.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Advanced Mathematics", found.Name)
}

func TestCategoryRepository_Delete(t *testing.T) {
	db := setupCategoryTestDB(t)
	repo := NewCategoryRepository(db)

	category := &model.Category{
		Name: "Science",
	}
	require.NoError(t, repo.Create(context.Background(), category))

	err := repo.Delete(context.Background(), category.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), category.ID)
	assert.Error(t, err)
}

func TestCategoryRepository_List(t *testing.T) {
	db := setupCategoryTestDB(t)
	repo := NewCategoryRepository(db)

	names := []string{"Mathematics", "Science", "Programming"}
	for _, name := range names {
		category := &model.Category{Name: name}
		require.NoError(t, repo.Create(context.Background(), category))
	}

	result, total, err := repo.List(context.Background(), &model.ListCategoriesRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Equal(t, 3, len(result))
}

func TestCategoryRepository_List_Pagination(t *testing.T) {
	db := setupCategoryTestDB(t)
	repo := NewCategoryRepository(db)

	names := []string{"Mathematics", "Science", "Programming"}
	for _, name := range names {
		category := &model.Category{Name: name}
		require.NoError(t, repo.Create(context.Background(), category))
	}

	result, total, err := repo.List(context.Background(), &model.ListCategoriesRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Equal(t, 2, len(result))
}

func TestCategoryRepository_List_DefaultPagination(t *testing.T) {
	db := setupCategoryTestDB(t)
	repo := NewCategoryRepository(db)

	result, _, err := repo.List(context.Background(), &model.ListCategoriesRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCategoryRepository_List_WithSearch(t *testing.T) {
	t.Skip("ILIKE is PostgreSQL-specific; this test requires PostgreSQL")
	db := setupCategoryTestDB(t)
	repo := NewCategoryRepository(db)

	for _, name := range []string{"Mathematics", "Science", "Programming"} {
		category := &model.Category{Name: name}
		require.NoError(t, repo.Create(context.Background(), category))
	}

	result, total, err := repo.List(context.Background(), &model.ListCategoriesRequest{
		Page:     1,
		PageSize: 20,
		Search:   "math",
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))
	assert.GreaterOrEqual(t, len(result), 1)
}

func TestNewCategoryRepository(t *testing.T) {
	db := setupCategoryTestDB(t)
	repo := NewCategoryRepository(db)
	assert.NotNil(t, repo)
}
