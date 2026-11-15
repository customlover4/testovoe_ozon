package service

import (
	"context"
	"errors"
	"fmt"
	"testovoe/internal/entities/post"
	"testovoe/pkg/errs"
)

type postService interface {
	NewPost(ctx context.Context, p post.Post) (post.Post, error)
	Posts(ctx context.Context, first int, after uint64) ([]post.Post, error)
	Post(ctx context.Context, id uint64) (post.Post, error)
	SetCommentsEnabled(ctx context.Context, postID uint64, enabled bool) error
}

func (s *service) NewPost(ctx context.Context, p post.Post) (post.Post, error) {
	const op = "internal.service.post.New"

	if p.Title == "" {
		return post.Post{}, errs.NewValidation("title", "empty field")
	}

	if p.Body == "" {
		return post.Post{}, errs.NewValidation("body", "empty field")
	}
	if p.AuthorID == 0 {
		return post.Post{}, errs.NewValidation("author_id", "invalid value")
	}

	created, err := s.stor.NewPost(ctx, p)
	if err != nil {
		return post.Post{}, fmt.Errorf("%s: %w", op, err)
	}

	return created, nil
}

func (s *service) SetCommentsEnabled(ctx context.Context, postID, authorID uint64, enabled bool) (post.Post, error) {
	const op = "internal.service.SetCommentsEnabled"

	if postID == 0 || authorID == 0 {
		return post.Post{}, errs.NewValidation("id", "invalid value")
	}

	p, err := s.Post(ctx, postID)
	if err != nil {
		return post.Post{}, err
	}
	if p.AuthorID != authorID {
		return post.Post{}, errs.ErrForbidden
	}

	if err := s.stor.SetCommentsEnabled(ctx, postID, enabled); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return post.Post{}, errs.ErrNotFound
		}
		return post.Post{}, fmt.Errorf("%s: %w", op, err)
	}
	p.CommentsEnabled = enabled
	return p, nil
}

func (s *service) Posts(ctx context.Context, first int, after uint64) ([]post.Post, uint64, error) {
	const op = "internal.service.Posts"

	if first <= 0 {
		first = 10
	}
	if first > 100 {
		first = 100
	}

	posts, err := s.stor.Posts(ctx, first+1, after)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	var next uint64
	if len(posts) > first {
		next = posts[first-1].ID
		posts = posts[:first]
	}
	return posts, next, nil
}

func (s *service) Post(ctx context.Context, id uint64) (post.Post, error) {
	const op = "internal.service.Post"

	if id < 1 {
		return post.Post{}, errs.NewValidation("id", "invalid value")
	}

	p, err := s.stor.Post(ctx, id)
	if errors.Is(err, errs.ErrNotFound) {
		return post.Post{}, errs.ErrNotFound
	} else if err != nil {
		return post.Post{}, fmt.Errorf("%s: %w", op, err)
	}

	return p, nil
}
