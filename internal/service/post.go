package service

import (
	"context"
	"fmt"
	"testovoe/internal/entities/post"
	"testovoe/pkg/errs"
)

type postService interface {
	NewPost(ctx context.Context, p post.Post) error
	Posts(ctx context.Context, first int, after uint64) ([]*post.Post, error)
}

func (s *service) NewPost(ctx context.Context, p post.Post) error {
	const op = "internal.service.post.New"

	if p.Title == "" {
		return errs.NewValidation("title", "empty field")
	}

	if p.Body == "" {
		return errs.NewValidation("body", "empty field")
	}

	err := s.stor.NewPost(ctx, p)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *service) Posts(ctx context.Context, first int, after uint64) ([]*post.Post, uint64, error) {
	const op = "internal.service.Posts"

	if first <= 0 {
		first = 10
	}
	if first > 100 {
		first = 100
	}

	posts, err := s.stor.Posts(ctx, first, after)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	var next uint64
	if len(posts) > 0 {
		next = posts[len(posts)-1].ID
	}
	return posts, next, nil
}
