package postgres

import (
	"context"
	"errors"
	"testing"

	"testovoe/internal/entities/post"
	"testovoe/pkg/errs"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
)

func TestNewPost(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		p := post.Post{
			Title:           "Title",
			Body:            "Body",
			AuthorID:        42,
			CommentsEnabled: true,
		}

		mock.ExpectQuery(`INSERT INTO posts`).
			WithArgs("Title", "Body", int64(42), true).
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(100)))

		created, err := store.NewPost(ctx, p)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created.ID != 100 {
			t.Fatalf("expected id 100, got %d", created.ID)
		}
		if created.Title != "Title" || created.Body != "Body" {
			t.Fatalf("unexpected post: %+v", created)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestPost(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectQuery(`SELECT id, title, body, author_id, comments_enabled FROM posts WHERE id = \$1`).
			WithArgs(int64(100)).
			WillReturnRows(pgxmock.NewRows([]string{"id", "title", "body", "author_id", "comments_enabled"}).
				AddRow(int64(100), "Title", "Body", int64(42), true))

		p, err := store.Post(ctx, 100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.ID != 100 || p.AuthorID != 42 || p.Title != "Title" {
			t.Fatalf("unexpected post: %+v", p)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectQuery(`SELECT id, title, body, author_id, comments_enabled FROM posts WHERE id = \$1`).
			WithArgs(int64(999)).
			WillReturnError(pgx.ErrNoRows)

		_, err := store.Post(ctx, 999)
		if !errors.Is(err, errs.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestPosts(t *testing.T) {
	ctx := context.Background()

	t.Run("empty", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectQuery(`SELECT id, title, body, author_id, comments_enabled FROM posts WHERE id > \$1 ORDER BY id LIMIT \$2`).
			WithArgs(int64(0), 10).
			WillReturnRows(pgxmock.NewRows([]string{"id", "title", "body", "author_id", "comments_enabled"}))

		posts, err := store.Posts(ctx, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(posts) != 0 {
			t.Fatalf("expected 0, got %d", len(posts))
		}
	})

	t.Run("success", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectQuery(`SELECT id, title, body, author_id, comments_enabled FROM posts WHERE id > \$1 ORDER BY id LIMIT \$2`).
			WithArgs(int64(0), 2).
			WillReturnRows(pgxmock.NewRows([]string{"id", "title", "body", "author_id", "comments_enabled"}).
				AddRow(int64(1), "T1", "B1", int64(10), true).
				AddRow(int64(2), "T2", "B2", int64(10), false))

		posts, err := store.Posts(ctx, 2, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(posts) != 2 {
			t.Fatalf("expected 2, got %d", len(posts))
		}
		if posts[0].ID != 1 || posts[1].ID != 2 {
			t.Fatalf("unexpected ids: %d, %d", posts[0].ID, posts[1].ID)
		}
		if !posts[0].CommentsEnabled || posts[1].CommentsEnabled {
			t.Fatalf("unexpected comments_enabled flags")
		}
	})

	t.Run("after cursor", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectQuery(`SELECT id, title, body, author_id, comments_enabled FROM posts WHERE id > \$1 ORDER BY id LIMIT \$2`).
			WithArgs(int64(5), 2).
			WillReturnRows(pgxmock.NewRows([]string{"id", "title", "body", "author_id", "comments_enabled"}).
				AddRow(int64(6), "T6", "B6", int64(10), true).
				AddRow(int64(7), "T7", "B7", int64(10), true))

		posts, err := store.Posts(ctx, 2, 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(posts) != 2 || posts[0].ID != 6 {
			t.Fatalf("unexpected posts: %+v", posts)
		}
	})
}
