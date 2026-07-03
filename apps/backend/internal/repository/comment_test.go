package repository_test

import (
	"context"
	"testing"

	"github.com/6sLOGAN78/go-protask/internal/model/comment"
	"github.com/6sLOGAN78/go-protask/internal/repository"
	testing_pkg "github.com/6sLOGAN78/go-protask/internal/testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommentRepository_AddComment(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	todoRepo := repository.NewTodoRepository(testServer)
	commentRepo := repository.NewCommentRepository(testServer)

	userID := uuid.New().String()
	testTodo := createTestTodo(t, ctx, todoRepo, userID)

	t.Run("add comment successfully", func(t *testing.T) {
		payload := &comment.AddCommentPayload{
			Content: "This is a test comment",
		}

		result, err := commentRepo.AddComment(ctx, userID, testTodo.ID, payload)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.NotEqual(t, uuid.Nil, result.ID)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, testTodo.ID, result.TodoID)
		assert.Equal(t, payload.Content, result.Content)
		assert.False(t, result.CreatedAt.IsZero())
		assert.False(t, result.UpdatedAt.IsZero())
	})

	t.Run("with canceled context", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		payload := &comment.AddCommentPayload{
			Content: "Canceled comment",
		}

		result, err := commentRepo.AddComment(canceledCtx, userID, testTodo.ID, payload)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCommentRepository_GetCommentsByTodoID(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	todoRepo := repository.NewTodoRepository(testServer)
	commentRepo := repository.NewCommentRepository(testServer)

	userID := uuid.New().String()
	testTodo := createTestTodo(t, ctx, todoRepo, userID)

	// Add a few comments
	payload1 := &comment.AddCommentPayload{Content: "First comment"}
	payload2 := &comment.AddCommentPayload{Content: "Second comment"}

	c1, err := commentRepo.AddComment(ctx, userID, testTodo.ID, payload1)
	require.NoError(t, err)
	c2, err := commentRepo.AddComment(ctx, userID, testTodo.ID, payload2)
	require.NoError(t, err)

	t.Run("get comments successfully", func(t *testing.T) {
		comments, err := commentRepo.GetCommentsByTodoID(ctx, userID, testTodo.ID)
		require.NoError(t, err)
		require.Len(t, comments, 2)

		assert.Equal(t, c1.ID, comments[0].ID)
		assert.Equal(t, c1.Content, comments[0].Content)
		assert.Equal(t, c2.ID, comments[1].ID)
		assert.Equal(t, c2.Content, comments[1].Content)
	})

	t.Run("get comments with wrong user id", func(t *testing.T) {
		wrongUserID := uuid.New().String()
		comments, err := commentRepo.GetCommentsByTodoID(ctx, wrongUserID, testTodo.ID)
		require.NoError(t, err)
		assert.Empty(t, comments)
	})

	t.Run("with canceled context", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		comments, err := commentRepo.GetCommentsByTodoID(canceledCtx, userID, testTodo.ID)
		assert.Error(t, err)
		assert.Nil(t, comments)
	})
}

func TestCommentRepository_GetCommentByID(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	todoRepo := repository.NewTodoRepository(testServer)
	commentRepo := repository.NewCommentRepository(testServer)

	userID := uuid.New().String()
	testTodo := createTestTodo(t, ctx, todoRepo, userID)

	payload := &comment.AddCommentPayload{Content: "Specific comment"}
	c, err := commentRepo.AddComment(ctx, userID, testTodo.ID, payload)
	require.NoError(t, err)

	t.Run("get comment by id successfully", func(t *testing.T) {
		result, err := commentRepo.GetCommentByID(ctx, userID, c.ID)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, c.ID, result.ID)
		assert.Equal(t, c.Content, result.Content)
		assert.Equal(t, testTodo.ID, result.TodoID)
	})

	t.Run("get non-existent comment", func(t *testing.T) {
		nonExistentID := uuid.New()
		result, err := commentRepo.GetCommentByID(ctx, userID, nonExistentID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("get comment with wrong user id", func(t *testing.T) {
		wrongUserID := uuid.New().String()
		result, err := commentRepo.GetCommentByID(ctx, wrongUserID, c.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("with canceled context", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		result, err := commentRepo.GetCommentByID(canceledCtx, userID, c.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCommentRepository_UpdateComment(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	todoRepo := repository.NewTodoRepository(testServer)
	commentRepo := repository.NewCommentRepository(testServer)

	userID := uuid.New().String()
	testTodo := createTestTodo(t, ctx, todoRepo, userID)

	payload := &comment.AddCommentPayload{Content: "Initial comment content"}
	c, err := commentRepo.AddComment(ctx, userID, testTodo.ID, payload)
	require.NoError(t, err)

	t.Run("update comment content successfully", func(t *testing.T) {
		newContent := "Updated comment content"
		result, err := commentRepo.UpdateComment(ctx, userID, c.ID, newContent)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, c.ID, result.ID)
		assert.Equal(t, newContent, result.Content)
		assert.True(t, result.UpdatedAt.After(c.UpdatedAt))
	})

	t.Run("update non-existent comment", func(t *testing.T) {
		nonExistentID := uuid.New()
		result, err := commentRepo.UpdateComment(ctx, userID, nonExistentID, "Updated text")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("update comment with wrong user id", func(t *testing.T) {
		wrongUserID := uuid.New().String()
		result, err := commentRepo.UpdateComment(ctx, wrongUserID, c.ID, "Updated text")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("with canceled context", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		result, err := commentRepo.UpdateComment(canceledCtx, userID, c.ID, "Updated text")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCommentRepository_DeleteComment(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	todoRepo := repository.NewTodoRepository(testServer)
	commentRepo := repository.NewCommentRepository(testServer)

	userID := uuid.New().String()
	testTodo := createTestTodo(t, ctx, todoRepo, userID)

	payload := &comment.AddCommentPayload{Content: "Comment to delete"}
	c, err := commentRepo.AddComment(ctx, userID, testTodo.ID, payload)
	require.NoError(t, err)

	t.Run("delete comment successfully", func(t *testing.T) {
		err := commentRepo.DeleteComment(ctx, userID, c.ID)
		require.NoError(t, err)

		// Verify comment is deleted
		result, err := commentRepo.GetCommentByID(ctx, userID, c.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("delete non-existent comment", func(t *testing.T) {
		nonExistentID := uuid.New()
		err := commentRepo.DeleteComment(ctx, userID, nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "comment not found")
	})

	t.Run("delete comment with wrong user id", func(t *testing.T) {
		wrongUserID := uuid.New().String()
		payloadTmp := &comment.AddCommentPayload{Content: "Comment with other user"}
		cTmp, err := commentRepo.AddComment(ctx, userID, testTodo.ID, payloadTmp)
		require.NoError(t, err)

		err = commentRepo.DeleteComment(ctx, wrongUserID, cTmp.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "comment not found")
	})

	t.Run("with canceled context", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		payloadTmp := &comment.AddCommentPayload{Content: "Canceled delete comment"}
		cTmp, err := commentRepo.AddComment(ctx, userID, testTodo.ID, payloadTmp)
		require.NoError(t, err)

		err = commentRepo.DeleteComment(canceledCtx, userID, cTmp.ID)
		assert.Error(t, err)
	})
}
