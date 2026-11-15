package local

import (
	"context"
	"errors"
	"testing"

	"testovoe/pkg/errs"
)

func TestNewUser(t *testing.T) {
	ctx := context.Background()

	t.Run("creates user", func(t *testing.T) {
		l := New()
		id, err := l.NewUser(ctx, "alice", "hash")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id == 0 {
			t.Fatal("expected non-zero id")
		}

		u, err := l.User(ctx, "alice")
		if err != nil {
			t.Fatalf("User: %v", err)
		}
		if u.ID != id || u.Login != "alice" || u.Password != "hash" {
			t.Fatalf("unexpected user: %+v", u)
		}
	})

	t.Run("duplicate login", func(t *testing.T) {
		l := New()
		if _, err := l.NewUser(ctx, "alice", "hash"); err != nil {
			t.Fatal(err)
		}
		_, err := l.NewUser(ctx, "alice", "hash2")
		if !errors.Is(err, errs.ErrNotUnique) {
			t.Fatalf("expected ErrNotUnique, got %v", err)
		}
	})

	t.Run("ids are unique and increasing", func(t *testing.T) {
		l := New()
		id1, _ := l.NewUser(ctx, "a", "h")
		id2, _ := l.NewUser(ctx, "b", "h")
		if id1 == id2 {
			t.Fatalf("expected unique ids, got %d and %d", id1, id2)
		}
		if id2 <= id1 {
			t.Fatalf("expected increasing ids, got %d then %d", id1, id2)
		}
	})
}

func TestUser(t *testing.T) {
	ctx := context.Background()

	t.Run("not found", func(t *testing.T) {
		l := New()
		_, err := l.User(ctx, "missing")
		if !errors.Is(err, errs.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("found by login", func(t *testing.T) {
		l := New()
		id, _ := l.NewUser(ctx, "alice", "hash")

		u, err := l.User(ctx, "alice")
		if err != nil {
			t.Fatalf("User: %v", err)
		}
		if u.ID != id {
			t.Fatalf("expected id %d, got %d", id, u.ID)
		}
	})
}
