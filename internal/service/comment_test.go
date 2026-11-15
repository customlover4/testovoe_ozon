package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testovoe/internal/entities/comment"
	"testovoe/internal/entities/post"
	"testovoe/pkg/errs"
)

func TestNewComment(t *testing.T) {
	ctx := context.Background()

	t.Run("empty text", func(t *testing.T) {
		svc := newTestService(&mockStorage{})
		_, err := svc.NewComment(ctx, comment.Comment{PostID: 1, AuthorID: 1, Text: ""})
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("text too long", func(t *testing.T) {
		svc := newTestService(&mockStorage{})
		longText := make([]rune, 2001)
		for i := range longText {
			longText[i] = 'a'
		}
		_, err := svc.NewComment(ctx, comment.Comment{PostID: 1, AuthorID: 1, Text: string(longText)})
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("post not found", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		storage.On("Post", ctx, uint64(1)).Return(post.Post{}, errs.ErrNotFound)

		_, err := svc.NewComment(ctx, comment.Comment{PostID: 1, AuthorID: 1, Text: "hello"})
		assert.True(t, errors.Is(err, errs.ErrNotFound))
	})

	t.Run("comments disabled", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		storage.On("Post", ctx, uint64(1)).Return(post.Post{ID: 1, CommentsEnabled: false}, nil)

		_, err := svc.NewComment(ctx, comment.Comment{PostID: 1, AuthorID: 1, Text: "hello"})
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("success", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		input := comment.Comment{PostID: 1, AuthorID: 1, Text: "hello"}
		expected := input
		expected.ID = 100

		storage.On("Post", ctx, uint64(1)).Return(post.Post{ID: 1, CommentsEnabled: true}, nil)
		storage.On("NewComment", ctx, input).Return(expected, nil)

		created, err := svc.NewComment(ctx, input)
		require.NoError(t, err)
		assert.Equal(t, uint64(100), created.ID)
	})
}

func TestComment(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid id", func(t *testing.T) {
		svc := newTestService(&mockStorage{})
		_, err := svc.Comment(ctx, 0)
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("not found", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		storage.On("Comment", ctx, uint64(10)).Return(comment.Comment{}, errs.ErrNotFound)

		_, err := svc.Comment(ctx, 10)
		assert.True(t, errors.Is(err, errs.ErrNotFound))
	})

	t.Run("success", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		expected := comment.Comment{ID: 10, Text: "test"}
		storage.On("Comment", ctx, uint64(10)).Return(expected, nil)

		c, err := svc.Comment(ctx, 10)
		require.NoError(t, err)
		assert.Equal(t, uint64(10), c.ID)
	})
}

func TestComments(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid post_id", func(t *testing.T) {
		svc := newTestService(&mockStorage{})
		_, _, err := svc.Comments(ctx, 0, 0, 10, 0)
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("post not found", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		storage.On("Post", ctx, uint64(1)).Return(post.Post{}, errs.ErrNotFound)

		_, _, err := svc.Comments(ctx, 1, 0, 10, 0)
		assert.True(t, errors.Is(err, errs.ErrNotFound))
	})

	t.Run("parent from another post", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		storage.On("Post", ctx, uint64(1)).Return(post.Post{ID: 1}, nil)
		storage.On("Comment", ctx, uint64(5)).Return(comment.Comment{ID: 5, PostID: 2}, nil)

		_, _, err := svc.Comments(ctx, 1, 5, 10, 0)
		assert.True(t, errors.Is(err, errs.ErrNotFound))
	})

	t.Run("success with pagination", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		storage.On("Post", ctx, uint64(1)).Return(post.Post{ID: 1}, nil)

		expectedComments := []comment.Comment{{ID: 10}, {ID: 11}}
		storage.On("Comments", ctx, uint64(1), uint64(0), 100, uint64(0)).Return(expectedComments, nil)

		comments, next, err := svc.Comments(ctx, 1, 0, 1000, 0)
		require.NoError(t, err)
		assert.Len(t, comments, 2)
		assert.Equal(t, uint64(11), next)
	})
}

func TestCommentTree(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid post_id", func(t *testing.T) {
		svc := newTestService(&mockStorage{})
		_, _, err := svc.CommentTree(ctx, 0, 10, 0)
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("post not found", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		storage.On("Post", ctx, uint64(1)).Return(post.Post{}, errs.ErrNotFound)

		_, _, err := svc.CommentTree(ctx, 1, 10, 0)
		assert.True(t, errors.Is(err, errs.ErrNotFound))
	})

	t.Run("success", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		storage.On("Post", ctx, uint64(1)).Return(post.Post{ID: 1}, nil)

		expectedComments := []comment.Comment{{ID: 10}}
		storage.On("CommentTree", ctx, uint64(1), 10, uint64(0)).Return(expectedComments, uint64(20), nil)

		comments, next, err := svc.CommentTree(ctx, 1, 10, 0)
		require.NoError(t, err)
		assert.Len(t, comments, 1)
		assert.Equal(t, uint64(20), next)
	})
}
