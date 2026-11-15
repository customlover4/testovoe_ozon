package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"testovoe/internal/entities/comment"
	"testovoe/pkg/errs"
)

func TestSetCommentsEnabled(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectExec(`UPDATE posts SET comments_enabled = \$2 WHERE id = \$1`).
			WithArgs(int64(1), true).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := store.SetCommentsEnabled(ctx, 1, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectExec(`UPDATE posts SET comments_enabled = \$2 WHERE id = \$1`).
			WithArgs(int64(99), true).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))

		err := store.SetCommentsEnabled(ctx, 99, true)
		if !errors.Is(err, errs.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestNewComment(t *testing.T) {
	ctx := context.Background()

	t.Run("success root comment", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO comments`).
			WithArgs(int64(1), pgxmock.AnyArg(), int64(1), "hello").
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(100)))
		mock.ExpectCommit()

		c, err := store.NewComment(ctx, comment.Comment{PostID: 1, ParentID: 0, AuthorID: 1, Text: "hello"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ID != 100 {
			t.Fatalf("expected id 100, got %d", c.ID)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("success reply", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT post_id FROM comments WHERE id = \$1 FOR KEY SHARE`).
			WithArgs(int64(42)).
			WillReturnRows(pgxmock.NewRows([]string{"post_id"}).AddRow(int64(1)))
		
		mock.ExpectQuery(`INSERT INTO comments`).
			WithArgs(int64(1), pgxmock.AnyArg(), int64(1), "reply").
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(101)))
		mock.ExpectCommit()

		c, err := store.NewComment(ctx, comment.Comment{PostID: 1, ParentID: 42, AuthorID: 1, Text: "reply"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ID != 101 {
			t.Fatalf("expected id 101, got %d", c.ID)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("parent not found", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT post_id FROM comments WHERE id = \$1 FOR KEY SHARE`).
			WithArgs(int64(999)).
			WillReturnError(pgx.ErrNoRows)
		mock.ExpectRollback()

		_, err := store.NewComment(ctx, comment.Comment{PostID: 1, ParentID: 999, AuthorID: 1, Text: "x"})
		if !errors.Is(err, errs.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("parent from another post", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT post_id FROM comments WHERE id = \$1 FOR KEY SHARE`).
			WithArgs(int64(42)).
			WillReturnRows(pgxmock.NewRows([]string{"post_id"}).AddRow(int64(2)))
		mock.ExpectRollback()

		_, err := store.NewComment(ctx, comment.Comment{PostID: 1, ParentID: 42, AuthorID: 1, Text: "x"})
		if !errors.Is(err, errs.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestComment(t *testing.T) {
	ctx := context.Background()

	t.Run("success root", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectQuery(`SELECT id, post_id, parent_id, author_id, text FROM comments WHERE id = \$1`).
			WithArgs(int64(10)).
			WillReturnRows(pgxmock.NewRows([]string{"id", "post_id", "parent_id", "author_id", "text"}).
				AddRow(int64(10), int64(1), (*int64)(nil), int64(5), "text"))

		c, err := store.Comment(ctx, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ID != 10 || c.ParentID != 0 || c.AuthorID != 5 {
			t.Fatalf("unexpected comment: %+v", c)
		}
	})

	t.Run("success reply", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		parentID := int64(10)
		mock.ExpectQuery(`SELECT id, post_id, parent_id, author_id, text FROM comments WHERE id = \$1`).
			WithArgs(int64(11)).
			WillReturnRows(pgxmock.NewRows([]string{"id", "post_id", "parent_id", "author_id", "text"}).
				AddRow(int64(11), int64(1), &parentID, int64(5), "reply"))

		c, err := store.Comment(ctx, 11)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ParentID != 10 {
			t.Fatalf("expected parent 10, got %d", c.ParentID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectQuery(`SELECT id, post_id, parent_id, author_id, text FROM comments WHERE id = \$1`).
			WithArgs(int64(999)).
			WillReturnError(pgx.ErrNoRows)

		_, err := store.Comment(ctx, 999)
		if !errors.Is(err, errs.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestComments(t *testing.T) {
	ctx := context.Background()

	t.Run("roots success", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectQuery(`SELECT id, post_id, parent_id, author_id, text FROM comments WHERE post_id = \$1 AND parent_id IS NULL AND id > \$2 ORDER BY id LIMIT \$3`).
			WithArgs(int64(1), int64(0), 10).
			WillReturnRows(pgxmock.NewRows([]string{"id", "post_id", "parent_id", "author_id", "text"}).
				AddRow(int64(10), int64(1), (*int64)(nil), int64(5), "root1").
				AddRow(int64(11), int64(1), (*int64)(nil), int64(5), "root2"))

		comments, err := store.Comments(ctx, 1, 0, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(comments) != 2 {
			t.Fatalf("expected 2, got %d", len(comments))
		}
	})

	t.Run("children success", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		parentID := int64(10)
		mock.ExpectQuery(`SELECT id, post_id, parent_id, author_id, text FROM comments WHERE post_id = \$1 AND parent_id = \$2 AND id > \$3 ORDER BY id LIMIT \$4`).
			WithArgs(int64(1), int64(10), int64(0), 10).
			WillReturnRows(pgxmock.NewRows([]string{"id", "post_id", "parent_id", "author_id", "text"}).
				AddRow(int64(11), int64(1), &parentID, int64(5), "child1"))

		comments, err := store.Comments(ctx, 1, 10, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(comments) != 1 || comments[0].ID != 11 {
			t.Fatalf("unexpected comments: %+v", comments)
		}
	})

	t.Run("after cursor not found (roots)", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments WHERE id = \$1 AND post_id = \$2 AND parent_id IS NULL\)`).
			WithArgs(int64(99), int64(1)).
			WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))

		_, err := store.Comments(ctx, 1, 0, 10, 99)
		if !errors.Is(err, errs.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("after cursor not found (children)", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments WHERE id = \$1 AND post_id = \$2 AND parent_id = \$3\)`).
			WithArgs(int64(99), int64(1), int64(10)).
			WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))

		_, err := store.Comments(ctx, 1, 10, 10, 99)
		if !errors.Is(err, errs.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestCommentTree(t *testing.T) {
	ctx := context.Background()

	t.Run("empty tree", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectQuery(`SELECT id FROM comments WHERE post_id = \$1 AND parent_id IS NULL AND id > \$2 ORDER BY id LIMIT \$3`).
			WithArgs(int64(1), int64(0), 11).
			WillReturnRows(pgxmock.NewRows([]string{"id"}))

		comments, next, err := store.CommentTree(ctx, 1, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(comments) != 0 || next != 0 {
			t.Fatalf("expected empty and next=0, got %d, %d", len(comments), next)
		}
	})

	t.Run("pagination with next cursor", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectQuery(`SELECT id FROM comments WHERE post_id = \$1 AND parent_id IS NULL AND id > \$2 ORDER BY id LIMIT \$3`).
			WithArgs(int64(1), int64(0), 3).
			WillReturnRows(pgxmock.NewRows([]string{"id"}).
				AddRow(int64(10)).AddRow(int64(20)).AddRow(int64(30)))

		parentID := int64(10)
		mock.ExpectQuery(`WITH RECURSIVE roots AS \(`).
			WithArgs([]int64{10, 20}).
			WillReturnRows(pgxmock.NewRows([]string{"id", "post_id", "parent_id", "author_id", "text"}).
				AddRow(int64(10), int64(1), (*int64)(nil), int64(5), "root1").
				AddRow(int64(11), int64(1), &parentID, int64(5), "child1").
				AddRow(int64(20), int64(1), (*int64)(nil), int64(5), "root2"))

		comments, next, err := store.CommentTree(ctx, 1, 2, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if next != 20 {
			t.Fatalf("expected next 20, got %d", next)
		}
		if len(comments) != 3 {
			t.Fatalf("expected 3, got %d", len(comments))
		}
		if comments[0].ID != 10 || comments[1].ID != 11 || comments[2].ID != 20 {
			t.Fatalf("unexpected order: %+v", comments)
		}
	})

	t.Run("after cursor not found", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments WHERE id = \$1 AND post_id = \$2 AND parent_id IS NULL\)`).
			WithArgs(int64(99), int64(1)).
			WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))

		_, _, err := store.CommentTree(ctx, 1, 10, 99)
		if !errors.Is(err, errs.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}