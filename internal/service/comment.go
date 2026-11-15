package service

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	"testovoe/internal/entities/comment"
	"testovoe/pkg/errs"
)

const maxCommentTextLength = 2000

type commentService interface {
	NewComment(ctx context.Context, c comment.Comment) (comment.Comment, error)
	Comment(ctx context.Context, id uint64) (comment.Comment, error)
	Comments(ctx context.Context, postID, parentID uint64, first int, after uint64) ([]comment.Comment, error)
	CommentTree(ctx context.Context, postID uint64, first int, after uint64) ([]comment.Comment, uint64, error)
}

func (s *service) NewComment(ctx context.Context, c comment.Comment) (comment.Comment, error) {
	const op = "internal.service.NewComment"

	if c.PostID == 0 {
		return comment.Comment{}, errs.NewValidation("post_id", "invalid value")
	}
	if c.AuthorID == 0 {
		return comment.Comment{}, errs.NewValidation("author_id", "invalid value")
	}
	if c.Text == "" {
		return comment.Comment{}, errs.NewValidation("text", "empty field")
	}
	if utf8.RuneCountInString(c.Text) > maxCommentTextLength {
		return comment.Comment{}, errs.NewValidation("text", "too long")
	}

	p, err := s.Post(ctx, c.PostID)
	if err != nil {
		return comment.Comment{}, err
	}
	if !p.CommentsEnabled {
		return comment.Comment{}, errs.NewValidation("post_id", "comments disabled")
	}

	created, err := s.stor.NewComment(ctx, c)
	if errors.Is(err, errs.ErrNotFound) {
		return comment.Comment{}, errs.ErrNotFound
	}
	if err != nil {
		return comment.Comment{}, fmt.Errorf("%s: %w", op, err)
	}

	return created, nil
}

func (s *service) Comment(ctx context.Context, id uint64) (comment.Comment, error) {
	const op = "internal.service.Comment"

	if id == 0 {
		return comment.Comment{}, errs.NewValidation("id", "invalid value")
	}

	c, err := s.stor.Comment(ctx, id)
	if errors.Is(err, errs.ErrNotFound) {
		return comment.Comment{}, errs.ErrNotFound
	}
	if err != nil {
		return comment.Comment{}, fmt.Errorf("%s: %w", op, err)
	}

	return c, nil
}

func (s *service) Comments(ctx context.Context, postID, parentID uint64, first int, after uint64) ([]comment.Comment, uint64, error) {
	const op = "internal.service.Comments"

	if postID == 0 {
		return nil, 0, errs.NewValidation("post_id", "invalid value")
	}

	if _, err := s.Post(ctx, postID); err != nil {
		return nil, 0, err
	}

	if parentID != 0 {
		parent, err := s.Comment(ctx, parentID)
		if err != nil {
			return nil, 0, err
		}
		if parent.PostID != postID {
			return nil, 0, errs.ErrNotFound
		}
	}

	if first <= 0 {
		first = 10
	}
	if first > 100 {
		first = 100
	}

	comments, err := s.stor.Comments(ctx, postID, parentID, first, after)
	if errors.Is(err, errs.ErrNotFound) {
		return nil, 0, errs.ErrNotFound
	}
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	var next uint64
	if len(comments) > 0 {
		next = comments[len(comments)-1].ID
	}

	return comments, next, nil
}

func (s *service) CommentTree(ctx context.Context, postID uint64, first int, after uint64) ([]comment.Comment, uint64, error) {
	const op = "internal.service.CommentTree"

	if postID == 0 {
		return nil, 0, errs.NewValidation("post_id", "invalid value")
	}
	if _, err := s.Post(ctx, postID); err != nil {
		return nil, 0, err
	}

	if first <= 0 {
		first = 10
	}
	if first > 100 {
		first = 100
	}

	comments, next, err := s.stor.CommentTree(ctx, postID, first, after)
	if errors.Is(err, errs.ErrNotFound) {
		return nil, 0, errs.ErrNotFound
	}
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	return comments, next, nil
}
