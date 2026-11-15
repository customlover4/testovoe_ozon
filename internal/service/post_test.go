package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testovoe/internal/entities/post"
	"testovoe/pkg/errs"
)

func TestNewPost(t *testing.T) {
	ctx := context.Background()

	t.Run("empty title", func(t *testing.T) {
		svc := newTestService(&mockStorage{})
		_, err := svc.NewPost(ctx, post.Post{Title: "", Body: "body", AuthorID: 1})
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("empty body", func(t *testing.T) {
		svc := newTestService(&mockStorage{})
		_, err := svc.NewPost(ctx, post.Post{Title: "title", Body: "", AuthorID: 1})
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("invalid author", func(t *testing.T) {
		svc := newTestService(&mockStorage{})
		_, err := svc.NewPost(ctx, post.Post{Title: "title", Body: "body", AuthorID: 0})
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("success", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		input := post.Post{Title: "title", Body: "body", AuthorID: 1}
		expected := input
		expected.ID = 100

		storage.On("NewPost", ctx, input).Return(expected, nil)

		created, err := svc.NewPost(ctx, input)
		require.NoError(t, err)
		assert.Equal(t, uint64(100), created.ID)
	})

	t.Run("storage error", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		input := post.Post{Title: "title", Body: "body", AuthorID: 1}
		storage.On("NewPost", ctx, input).Return(post.Post{}, errors.New("db error"))

		_, err := svc.NewPost(ctx, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "internal.service.post.New")
	})
}

func TestSetCommentsEnabled(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid ids", func(t *testing.T) {
		svc := newTestService(&mockStorage{})
		_, err := svc.SetCommentsEnabled(ctx, 0, 1, true)
		assert.True(t, errors.Is(err, errs.ErrValidation))

		_, err = svc.SetCommentsEnabled(ctx, 1, 0, true)
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("post not found", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		storage.On("Post", ctx, uint64(1)).Return(post.Post{}, errs.ErrNotFound)

		_, err := svc.SetCommentsEnabled(ctx, 1, 1, true)
		assert.True(t, errors.Is(err, errs.ErrNotFound))
	})

	t.Run("forbidden (not author)", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		storage.On("Post", ctx, uint64(1)).Return(post.Post{ID: 1, AuthorID: 2}, nil)

		_, err := svc.SetCommentsEnabled(ctx, 1, 1, true)
		assert.True(t, errors.Is(err, errs.ErrForbidden))
	})

	t.Run("success", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		storage.On("Post", ctx, uint64(1)).Return(post.Post{ID: 1, AuthorID: 1, CommentsEnabled: true}, nil)
		storage.On("SetCommentsEnabled", ctx, uint64(1), false).Return(nil)

		updated, err := svc.SetCommentsEnabled(ctx, 1, 1, false)
		require.NoError(t, err)
		assert.False(t, updated.CommentsEnabled)
	})

	t.Run("storage not found", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		storage.On("Post", ctx, uint64(1)).Return(post.Post{ID: 1, AuthorID: 1}, nil)
		storage.On("SetCommentsEnabled", ctx, uint64(1), false).Return(errs.ErrNotFound)

		_, err := svc.SetCommentsEnabled(ctx, 1, 1, false)
		assert.True(t, errors.Is(err, errs.ErrNotFound))
	})
}

func TestPosts(t *testing.T) {
	ctx := context.Background()

	t.Run("default pagination", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		storage.On("Posts", ctx, 11, uint64(0)).Return([]post.Post{}, nil)

		posts, next, err := svc.Posts(ctx, 0, 0)
		require.NoError(t, err)
		assert.Len(t, posts, 0)
		assert.Equal(t, uint64(0), next)
	})

	t.Run("max pagination", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		storage.On("Posts", ctx, 101, uint64(0)).Return([]post.Post{}, nil)

		posts, next, err := svc.Posts(ctx, 1000, 0)
		require.NoError(t, err)
		assert.Len(t, posts, 0)
		assert.Equal(t, uint64(0), next)
	})

	t.Run("has next page", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		storage.On("Posts", ctx, 3, uint64(0)).Return([]post.Post{
			{ID: 1}, {ID: 2}, {ID: 3},
		}, nil)

		posts, next, err := svc.Posts(ctx, 2, 0)
		require.NoError(t, err)
		assert.Len(t, posts, 2) 
		assert.Equal(t, uint64(2), next)
	})

	t.Run("no next page", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		storage.On("Posts", ctx, 3, uint64(0)).Return([]post.Post{
			{ID: 1}, {ID: 2},
		}, nil)

		posts, next, err := svc.Posts(ctx, 2, 0)
		require.NoError(t, err)
		assert.Len(t, posts, 2)
		assert.Equal(t, uint64(0), next)
	})
}

func TestPost(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid id", func(t *testing.T) {
		svc := newTestService(&mockStorage{})
		_, err := svc.Post(ctx, 0)
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("not found", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		storage.On("Post", ctx, uint64(1)).Return(post.Post{}, errs.ErrNotFound)

		_, err := svc.Post(ctx, 1)
		assert.True(t, errors.Is(err, errs.ErrNotFound))
	})

	t.Run("success", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newTestService(storage)

		expected := post.Post{ID: 1, Title: "test"}
		storage.On("Post", ctx, uint64(1)).Return(expected, nil)

		p, err := svc.Post(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), p.ID)
	})
}
