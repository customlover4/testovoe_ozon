package postgres

import (
	"context"
	"errors"
	"fmt"
	"testovoe/internal/entities/comment"
	"testovoe/pkg/errs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Store) SetCommentsEnabled(ctx context.Context, postID uint64, enabled bool) error {
	result, err := s.pool.Exec(ctx, `UPDATE posts SET comments_enabled = $2 WHERE id = $1`, int64(postID), enabled)
	if err != nil {
		return fmt.Errorf("update post comments setting: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (s *Store) NewComment(ctx context.Context, c comment.Comment) (comment.Comment, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return comment.Comment{}, fmt.Errorf("begin comment transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if c.ParentID != 0 {
		var parentPostID int64
		err := tx.QueryRow(ctx, `SELECT post_id FROM comments WHERE id = $1 FOR KEY SHARE`, int64(c.ParentID)).Scan(&parentPostID)
		if errors.Is(err, pgx.ErrNoRows) || uint64(parentPostID) != c.PostID {
			return comment.Comment{}, errs.ErrNotFound
		}
		if err != nil {
			return comment.Comment{}, fmt.Errorf("select parent comment: %w", err)
		}
	}

	var id int64
	var parentID *int64
	if c.ParentID != 0 {
		value := int64(c.ParentID)
		parentID = &value
	}
	err = tx.QueryRow(ctx, `
INSERT INTO comments (post_id, parent_id, author_id, text)
VALUES ($1, $2, $3, $4) RETURNING id`, int64(c.PostID), parentID, int64(c.AuthorID), c.Text).Scan(&id)
	if err != nil {
		return comment.Comment{}, fmt.Errorf("insert comment: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return comment.Comment{}, fmt.Errorf("commit comment transaction: %w", err)
	}
	c.ID = uint64(id)
	return c, nil
}

func (s *Store) Comment(ctx context.Context, id uint64) (comment.Comment, error) {
	var c comment.Comment
	var commentID, postID, authorID int64
	var parentID *int64
	err := s.pool.QueryRow(ctx, `
SELECT id, post_id, parent_id, author_id, text FROM comments WHERE id = $1`, int64(id)).
		Scan(&commentID, &postID, &parentID, &authorID, &c.Text)
	if errors.Is(err, pgx.ErrNoRows) {
		return comment.Comment{}, errs.ErrNotFound
	}
	if err != nil {
		return comment.Comment{}, fmt.Errorf("select comment: %w", err)
	}
	c.ID, c.PostID, c.AuthorID = uint64(commentID), uint64(postID), uint64(authorID)
	if parentID != nil {
		c.ParentID = uint64(*parentID)
	}
	return c, nil
}

func (s *Store) Comments(ctx context.Context, postID, parentID uint64, first int, after uint64) ([]comment.Comment, error) {
	if after != 0 {
		var found bool
		var err error
		if parentID == 0 {
			err = s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1 AND post_id = $2 AND parent_id IS NULL)`, int64(after), int64(postID)).Scan(&found)
		} else {
			err = s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1 AND post_id = $2 AND parent_id = $3)`, int64(after), int64(postID), int64(parentID)).Scan(&found)
		}
		if err != nil {
			return nil, fmt.Errorf("validate comment cursor: %w", err)
		}
		if !found {
			return nil, errs.ErrNotFound
		}
	}

	query := `SELECT id, post_id, parent_id, author_id, text FROM comments WHERE post_id = $1 AND parent_id IS NULL AND id > $2 ORDER BY id LIMIT $3`
	args := []any{int64(postID), int64(after), first}
	if parentID != 0 {
		query = `SELECT id, post_id, parent_id, author_id, text FROM comments WHERE post_id = $1 AND parent_id = $2 AND id > $3 ORDER BY id LIMIT $4`
		args = []any{int64(postID), int64(parentID), int64(after), first}
	}
	return s.queryComments(ctx, query, args...)
}

func (s *Store) CommentTree(ctx context.Context, postID uint64, first int, after uint64) ([]comment.Comment, uint64, error) {
	if after != 0 {
		var found bool
		err := s.pool.QueryRow(ctx, `
SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1 AND post_id = $2 AND parent_id IS NULL)`, int64(after), int64(postID)).Scan(&found)
		if err != nil {
			return nil, 0, fmt.Errorf("validate comment tree cursor: %w", err)
		}
		if !found {
			return nil, 0, errs.ErrNotFound
		}
	}
	if first <= 0 {
		first = 10
	}

	rows, err := s.pool.Query(ctx, `
SELECT id FROM comments
WHERE post_id = $1 AND parent_id IS NULL AND id > $2
ORDER BY id LIMIT $3`, int64(postID), int64(after), first+1)
	if err != nil {
		return nil, 0, fmt.Errorf("select comment roots: %w", err)
	}
	rootIDs := make([]int64, 0, first+1)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, 0, fmt.Errorf("scan comment root: %w", err)
		}
		rootIDs = append(rootIDs, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, 0, fmt.Errorf("iterate comment roots: %w", err)
	}
	rows.Close()

	var next uint64
	if len(rootIDs) > first {
		next = uint64(rootIDs[first-1])
		rootIDs = rootIDs[:first]
	}
	if len(rootIDs) == 0 {
		return []comment.Comment{}, next, nil
	}

	const query = `
WITH RECURSIVE roots AS (
    SELECT id, root_order
    FROM unnest($1::bigint[]) WITH ORDINALITY AS root(id, root_order)
), tree AS (
    SELECT c.id, c.post_id, c.parent_id, c.author_id, c.text, roots.root_order, ARRAY[c.id] AS path
    FROM comments c JOIN roots ON roots.id = c.id
    UNION ALL
    SELECT c.id, c.post_id, c.parent_id, c.author_id, c.text, tree.root_order, tree.path || c.id
    FROM comments c JOIN tree ON c.parent_id = tree.id
)
SELECT id, post_id, parent_id, author_id, text FROM tree ORDER BY root_order, path`
	comments, err := s.queryComments(ctx, query, rootIDs)
	if err != nil {
		return nil, 0, err
	}
	return comments, next, nil
}

func (s *Store) queryComments(ctx context.Context, query string, args ...any) ([]comment.Comment, error) {
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select comments: %w", err)
	}
	defer rows.Close()

	comments := make([]comment.Comment, 0)
	for rows.Next() {
		var c comment.Comment
		var id, postID, authorID int64
		var parentID *int64
		if err := rows.Scan(&id, &postID, &parentID, &authorID, &c.Text); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		c.ID, c.PostID, c.AuthorID = uint64(id), uint64(postID), uint64(authorID)
		if parentID != nil {
			c.ParentID = uint64(*parentID)
		}
		comments = append(comments, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate comments: %w", err)
	}
	return comments, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
