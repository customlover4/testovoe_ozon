package storage

import (
	"context"
	"fmt"
	"strings"

	"testovoe/internal/entities/comment"
	"testovoe/internal/entities/post"
	"testovoe/internal/entities/user"
	"testovoe/internal/storage/local"
	"testovoe/internal/storage/postgres"
)

type comms interface {
	CommentTree(ctx context.Context, postID uint64, first int, after uint64) ([]comment.Comment, uint64, error)
	Comments(ctx context.Context, postID, parentID uint64, first int, after uint64) ([]comment.Comment, error)
	Comment(ctx context.Context, id uint64) (comment.Comment, error)
	NewComment(ctx context.Context, c comment.Comment) (comment.Comment, error)
}

type posts interface {
	NewPost(ctx context.Context, p post.Post) (post.Post, error)
	Post(ctx context.Context, id uint64) (post.Post, error)
	Posts(ctx context.Context, first int, after uint64) ([]post.Post, error)
	SetCommentsEnabled(ctx context.Context, postID uint64, enabled bool) error
}

type users interface {
	NewUser(ctx context.Context, login, password string) (uint64, error)
	User(ctx context.Context, login string) (user.User, error)
}

type Storage interface {
	comms
	posts
	users
}

func New(ctx context.Context, storageType, connString string) (Storage, func(), error) {
	switch strings.ToLower(storageType) {
	case "", "memory", "local":
		return local.New(), nil, nil
	case "postgres", "postgresql":
		store, err := postgres.New(ctx, connString)
		if err != nil {
			return nil, nil, fmt.Errorf("connect postgres: %w", err)
		}
		return store, store.Close, nil
	default:
		return nil, nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}
