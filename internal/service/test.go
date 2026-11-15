package service

import (
	"context"
	"testovoe/internal/entities/comment"
	"testovoe/internal/entities/post"
	"testovoe/internal/entities/user"

	"github.com/stretchr/testify/mock"
)

type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) NewUser(ctx context.Context, login, password string) (uint64, error) {
	args := m.Called(ctx, login, password)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *mockStorage) User(ctx context.Context, login string) (user.User, error) {
	args := m.Called(ctx, login)
	return args.Get(0).(user.User), args.Error(1)
}

func (m *mockStorage) NewPost(ctx context.Context, p post.Post) (post.Post, error) {
	args := m.Called(ctx, p)
	return args.Get(0).(post.Post), args.Error(1)
}

func (m *mockStorage) Post(ctx context.Context, id uint64) (post.Post, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(post.Post), args.Error(1)
}

func (m *mockStorage) Posts(ctx context.Context, first int, after uint64) ([]post.Post, error) {
	args := m.Called(ctx, first, after)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]post.Post), args.Error(1)
}

func (m *mockStorage) SetCommentsEnabled(ctx context.Context, postID uint64, enabled bool) error {
	args := m.Called(ctx, postID, enabled)
	return args.Error(0)
}

func (m *mockStorage) NewComment(ctx context.Context, c comment.Comment) (comment.Comment, error) {
	args := m.Called(ctx, c)
	return args.Get(0).(comment.Comment), args.Error(1)
}

func (m *mockStorage) Comment(ctx context.Context, id uint64) (comment.Comment, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(comment.Comment), args.Error(1)
}

func (m *mockStorage) Comments(ctx context.Context, postID, parentID uint64, first int, after uint64) ([]comment.Comment, error) {
	args := m.Called(ctx, postID, parentID, first, after)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]comment.Comment), args.Error(1)
}

func (m *mockStorage) CommentTree(ctx context.Context, postID uint64, first int, after uint64) ([]comment.Comment, uint64, error) {
	args := m.Called(ctx, postID, first, after)
	if args.Get(0) == nil {
		return nil, args.Get(1).(uint64), args.Error(2)
	}
	return args.Get(0).([]comment.Comment), args.Get(1).(uint64), args.Error(2)
}

func newTestService(storage *mockStorage) *service {
	return &service{stor: storage}
}
