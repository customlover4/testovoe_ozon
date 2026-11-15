package postgres

import (
	"context"
	"errors"
	"fmt"
	"testovoe/internal/entities/post"
	"testovoe/pkg/errs"

	"github.com/jackc/pgx/v5"
)

func (s *Store) NewPost(ctx context.Context, p post.Post) (post.Post, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
INSERT INTO posts (title, body, author_id, comments_enabled)
VALUES ($1, $2, $3, $4) RETURNING id`, p.Title, p.Body, int64(p.AuthorID), p.CommentsEnabled).Scan(&id)
	if err != nil {
		return post.Post{}, fmt.Errorf("insert post: %w", err)
	}
	p.ID = uint64(id)
	return p, nil
}

func (s *Store) Post(ctx context.Context, id uint64) (post.Post, error) {
	var p post.Post
	var postID, authorID int64
	err := s.pool.QueryRow(ctx, `
SELECT id, title, body, author_id, comments_enabled FROM posts WHERE id = $1`, int64(id)).
		Scan(&postID, &p.Title, &p.Body, &authorID, &p.CommentsEnabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return post.Post{}, errs.ErrNotFound
	}
	if err != nil {
		return post.Post{}, fmt.Errorf("select post: %w", err)
	}
	p.ID, p.AuthorID = uint64(postID), uint64(authorID)
	return p, nil
}

func (s *Store) Posts(ctx context.Context, first int, after uint64) ([]post.Post, error) {
	rows, err := s.pool.Query(ctx, `
SELECT id, title, body, author_id, comments_enabled
FROM posts WHERE id > $1 ORDER BY id LIMIT $2`, int64(after), first)
	if err != nil {
		return nil, fmt.Errorf("select posts: %w", err)
	}
	defer rows.Close()

	posts := make([]post.Post, 0, first)
	for rows.Next() {
		var p post.Post
		var id, authorID int64
		if err := rows.Scan(&id, &p.Title, &p.Body, &authorID, &p.CommentsEnabled); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		p.ID, p.AuthorID = uint64(id), uint64(authorID)
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate posts: %w", err)
	}
	return posts, nil
}
