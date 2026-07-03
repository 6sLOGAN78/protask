package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/6sLOGAN78/go-protask/internal/model/category"
	"github.com/6sLOGAN78/go-protask/internal/repository"
	testing_pkg "github.com/6sLOGAN78/go-protask/internal/testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoryRepository_CreateCategory(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	categoryRepo := repository.NewCategoryRepository(testServer)

	userID := uuid.New().String()

	t.Run("create category successfully", func(t *testing.T) {
		payload := &category.CreateCategoryPayload{
			Name:        "Work",
			Color:       "#ff0000",
			Description: testing_pkg.Ptr("Work related categories"),
		}

		result, err := categoryRepo.CreateCategory(ctx, userID, payload)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.NotEqual(t, uuid.Nil, result.ID)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, payload.Name, result.Name)
		assert.Equal(t, payload.Color, result.Color)
		assert.Equal(t, payload.Description, result.Description)
		assert.False(t, result.CreatedAt.IsZero())
		assert.False(t, result.UpdatedAt.IsZero())
	})

	t.Run("create category with minimum fields", func(t *testing.T) {
		payload := &category.CreateCategoryPayload{
			Name:  "Personal",
			Color: "#00ff00",
		}

		result, err := categoryRepo.CreateCategory(ctx, userID, payload)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, payload.Name, result.Name)
		assert.Equal(t, payload.Color, result.Color)
		assert.Nil(t, result.Description)
	})

	t.Run("with canceled context", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		payload := &category.CreateCategoryPayload{
			Name:  "Canceled",
			Color: "#0000ff",
		}

		result, err := categoryRepo.CreateCategory(canceledCtx, userID, payload)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCategoryRepository_GetCategoryByID(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	categoryRepo := repository.NewCategoryRepository(testServer)

	userID := uuid.New().String()
	c := createTestCategory(t, ctx, categoryRepo, userID, "Work", "#ff0000")

	t.Run("get category successfully", func(t *testing.T) {
		result, err := categoryRepo.GetCategoryByID(ctx, userID, c.ID)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, c.ID, result.ID)
		assert.Equal(t, c.Name, result.Name)
		assert.Equal(t, c.Color, result.Color)
	})

	t.Run("get non-existent category", func(t *testing.T) {
		nonExistentID := uuid.New()
		result, err := categoryRepo.GetCategoryByID(ctx, userID, nonExistentID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("get category with wrong user id", func(t *testing.T) {
		wrongUserID := uuid.New().String()
		result, err := categoryRepo.GetCategoryByID(ctx, wrongUserID, c.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("with canceled context", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		result, err := categoryRepo.GetCategoryByID(canceledCtx, userID, c.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCategoryRepository_GetCategories(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	categoryRepo := repository.NewCategoryRepository(testServer)

	userID := uuid.New().String()

	// Create a few categories
	c1 := createTestCategory(t, ctx, categoryRepo, userID, "Work", "#ff0000")
	time.Sleep(10 * time.Millisecond) // separate timestamps
	c2 := createTestCategory(t, ctx, categoryRepo, userID, "Urgent", "#00ff00")
	time.Sleep(10 * time.Millisecond)
	c3 := createTestCategory(t, ctx, categoryRepo, userID, "Personal", "#0000ff")

	t.Run("get categories default pagination (alphabetical name asc)", func(t *testing.T) {
		page := 1
		limit := 10
		query := &category.GetCategoriesQuery{
			Page:  &page,
			Limit: &limit,
		}
		require.NoError(t, query.Validate())

		result, err := categoryRepo.GetCategories(ctx, userID, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Len(t, result.Data, 3)
		assert.Equal(t, page, result.Page)
		assert.Equal(t, limit, result.Limit)
		assert.Equal(t, 3, result.Total)
		assert.Equal(t, 1, result.TotalPages)

		// Verification in default alphabetical sort: Personal, Urgent, Work
		assert.Equal(t, c3.ID, result.Data[0].ID) // Personal
		assert.Equal(t, c2.ID, result.Data[1].ID) // Urgent
		assert.Equal(t, c1.ID, result.Data[2].ID) // Work
	})

	t.Run("get categories sorted by created_at desc", func(t *testing.T) {
		page := 1
		limit := 10
		sort := "created_at"
		order := "desc"
		query := &category.GetCategoriesQuery{
			Page:  &page,
			Limit: &limit,
			Sort:  &sort,
			Order: &order,
		}
		require.NoError(t, query.Validate())

		result, err := categoryRepo.GetCategories(ctx, userID, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Len(t, result.Data, 3)
		// Created newest first: Personal (c3), Urgent (c2), Work (c1)
		assert.Equal(t, c3.ID, result.Data[0].ID)
		assert.Equal(t, c2.ID, result.Data[1].ID)
		assert.Equal(t, c1.ID, result.Data[2].ID)
	})

	t.Run("search categories by name", func(t *testing.T) {
		page := 1
		limit := 10
		search := "e" // Matches "Urgent" and "Personal"
		query := &category.GetCategoriesQuery{
			Page:   &page,
			Limit:  &limit,
			Search: &search,
		}
		require.NoError(t, query.Validate())

		result, err := categoryRepo.GetCategories(ctx, userID, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Len(t, result.Data, 2)
		assert.Equal(t, 2, result.Total)
		// Default name sort: Personal, Urgent
		assert.Equal(t, c3.ID, result.Data[0].ID) // Personal
		assert.Equal(t, c2.ID, result.Data[1].ID) // Urgent
	})

	t.Run("with canceled context", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		page := 1
		limit := 10
		query := &category.GetCategoriesQuery{
			Page:  &page,
			Limit: &limit,
		}
		require.NoError(t, query.Validate())

		result, err := categoryRepo.GetCategories(canceledCtx, userID, query)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCategoryRepository_UpdateCategory(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	categoryRepo := repository.NewCategoryRepository(testServer)

	userID := uuid.New().String()
	c := createTestCategory(t, ctx, categoryRepo, userID, "Work", "#ff0000")

	t.Run("update category fields successfully", func(t *testing.T) {
		newName := "Professional"
		newColor := "#ffffff"
		newDesc := "Professional items only"
		payload := &category.UpdateCategoryPayload{
			Name:        &newName,
			Color:       &newColor,
			Description: &newDesc,
		}

		result, err := categoryRepo.UpdateCategory(ctx, userID, c.ID, payload)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, c.ID, result.ID)
		assert.Equal(t, newName, result.Name)
		assert.Equal(t, newColor, result.Color)
		assert.Equal(t, newDesc, *result.Description)
		assert.True(t, result.UpdatedAt.After(c.UpdatedAt))
	})

	t.Run("update non-existent category", func(t *testing.T) {
		nonExistentID := uuid.New()
		newName := "Error Category"
		payload := &category.UpdateCategoryPayload{
			Name: &newName,
		}
		result, err := categoryRepo.UpdateCategory(ctx, userID, nonExistentID, payload)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("with canceled context", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		newName := "Cancel Category"
		payload := &category.UpdateCategoryPayload{
			Name: &newName,
		}
		result, err := categoryRepo.UpdateCategory(canceledCtx, userID, c.ID, payload)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCategoryRepository_DeleteCategory(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	categoryRepo := repository.NewCategoryRepository(testServer)

	userID := uuid.New().String()
	c := createTestCategory(t, ctx, categoryRepo, userID, "Delete Me", "#999999")

	t.Run("delete category successfully", func(t *testing.T) {
		err := categoryRepo.DeleteCategory(ctx, userID, c.ID)
		require.NoError(t, err)

		// Verify category is deleted
		result, err := categoryRepo.GetCategoryByID(ctx, userID, c.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("delete non-existent category", func(t *testing.T) {
		nonExistentID := uuid.New()
		err := categoryRepo.DeleteCategory(ctx, userID, nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "category not found")
	})

	t.Run("delete category with wrong user id", func(t *testing.T) {
		wrongUserID := uuid.New().String()
		cTmp := createTestCategory(t, ctx, categoryRepo, userID, "Temp Category", "#123456")

		err := categoryRepo.DeleteCategory(ctx, wrongUserID, cTmp.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "category not found")
	})

	t.Run("with canceled context", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		cTmp := createTestCategory(t, ctx, categoryRepo, userID, "Canceled Delete", "#234567")

		err := categoryRepo.DeleteCategory(canceledCtx, userID, cTmp.ID)
		assert.Error(t, err)
	})
}

func createTestCategory(t *testing.T, ctx context.Context, repo *repository.CategoryRepository, userID, name, color string) *category.Category {
	t.Helper()

	payload := &category.CreateCategoryPayload{
		Name:  name,
		Color: color,
	}

	result, err := repo.CreateCategory(ctx, userID, payload)
	require.NoError(t, err)

	return result
}
