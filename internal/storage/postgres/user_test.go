package postgres

import (
	"context"
	"errors"
	"testing"

	"testovoe/pkg/errs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v5"
)

func TestNewUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectQuery(`INSERT INTO users \(login, password\) VALUES \(\$1, \$2\) RETURNING id`).
			WithArgs("alice", "hashed_password").
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(1)))

		id, err := store.NewUser(ctx, "alice", "hashed_password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != 1 {
			t.Fatalf("expected id 1, got %d", id)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("unique violation", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		pgErr := &pgconn.PgError{Code: "23505"}

		mock.ExpectQuery(`INSERT INTO users \(login, password\) VALUES \(\$1, \$2\) RETURNING id`).
			WithArgs("alice", "hashed_password").
			WillReturnError(pgErr)

		_, err := store.NewUser(ctx, "alice", "hashed_password")
		if !errors.Is(err, errs.ErrNotUnique) {
			t.Fatalf("expected ErrNotUnique, got %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		dbErr := errors.New("connection lost")

		mock.ExpectQuery(`INSERT INTO users \(login, password\) VALUES \(\$1, \$2\) RETURNING id`).
			WithArgs("bob", "hashed_password").
			WillReturnError(dbErr)

		_, err := store.NewUser(ctx, "bob", "hashed_password")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if errors.Is(err, errs.ErrNotUnique) {
			t.Fatal("did not expect ErrNotUnique")
		}
		if !contains(err.Error(), "insert user") {
			t.Fatalf("expected wrapped error, got %v", err)
		}
	})
}

func TestUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectQuery(`SELECT id, login, password FROM users WHERE login = \$1`).
			WithArgs("alice").
			WillReturnRows(pgxmock.NewRows([]string{"id", "login", "password"}).
				AddRow(int64(1), "alice", "hashed_password"))

		u, err := store.User(ctx, "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.ID != 1 {
			t.Fatalf("expected id 1, got %d", u.ID)
		}
		if u.Login != "alice" {
			t.Fatalf("expected login 'alice', got '%s'", u.Login)
		}
		if u.Password != "hashed_password" {
			t.Fatalf("expected password 'hashed_password', got '%s'", u.Password)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		mock.ExpectQuery(`SELECT id, login, password FROM users WHERE login = \$1`).
			WithArgs("unknown").
			WillReturnError(pgx.ErrNoRows)

		_, err := store.User(ctx, "unknown")
		if !errors.Is(err, errs.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		store := &Store{pool: mock}

		dbErr := errors.New("connection lost")

		mock.ExpectQuery(`SELECT id, login, password FROM users WHERE login = \$1`).
			WithArgs("alice").
			WillReturnError(dbErr)

		_, err := store.User(ctx, "alice")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if errors.Is(err, errs.ErrNotFound) {
			t.Fatal("did not expect ErrNotFound")
		}
		if !contains(err.Error(), "select user") {
			t.Fatalf("expected wrapped error, got %v", err)
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
