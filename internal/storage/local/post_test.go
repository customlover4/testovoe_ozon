package local

import (
	"context"
	"errors"
	"testing"

	"testovoe/internal/entities/post"
	"testovoe/pkg/errs"
)

func TestNewPost(t *testing.T) {
	ctx := context.Background()

	t.Run("creates post", func(t *testing.T) {
		l := New()
		p, err := l.NewPost(ctx, post.Post{Title: "T", Body: "B", AuthorID: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.ID == 0 {
			t.Fatal("expected non-zero id")
		}

		got, err := l.Post(ctx, p.ID)
		if err != nil {
			t.Fatalf("Post: %v", err)
		}
		if got.Title != "T" || got.Body != "B" || got.AuthorID != 1 {
			t.Fatalf("unexpected post: %+v", got)
		}
	})

	t.Run("ids are unique", func(t *testing.T) {
		l := New()
		p1, _ := l.NewPost(ctx, post.Post{Title: "A"})
		p2, _ := l.NewPost(ctx, post.Post{Title: "B"})
		if p1.ID == p2.ID {
			t.Fatalf("expected unique ids, got %d", p1.ID)
		}
	})
}

func TestPost(t *testing.T) {
	ctx := context.Background()

	t.Run("not found", func(t *testing.T) {
		l := New()
		_, err := l.Post(ctx, 999)
		if !errors.Is(err, errs.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestPosts(t *testing.T) {
	ctx := context.Background()

	t.Run("empty", func(t *testing.T) {
		l := New()
		posts, err := l.Posts(ctx, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(posts) != 0 {
			t.Fatalf("expected empty, got %d", len(posts))
		}
	})

	t.Run("first page", func(t *testing.T) {
		l := New()
		for i := 0; i < 5; i++ {
			if _, err := l.NewPost(ctx, post.Post{Title: "T"}); err != nil {
				t.Fatal(err)
			}
		}

		posts, err := l.Posts(ctx, 2, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(posts) != 2 {
			t.Fatalf("expected 2, got %d", len(posts))
		}
	})

	t.Run("after cursor", func(t *testing.T) {
		l := New()
		var ids []uint64
		for i := 0; i < 5; i++ {
			p, _ := l.NewPost(ctx, post.Post{Title: "T"})
			ids = append(ids, p.ID)
		}

		posts, err := l.Posts(ctx, 2, ids[1]) // после второго
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(posts) != 2 {
			t.Fatalf("expected 2, got %d", len(posts))
		}
		if posts[0].ID != ids[2] {
			t.Fatalf("expected first post id %d, got %d", ids[2], posts[0].ID)
		}
	})

	t.Run("last page shorter", func(t *testing.T) {
		l := New()
		var ids []uint64
		for i := 0; i < 3; i++ {
			p, _ := l.NewPost(ctx, post.Post{Title: "T"})
			ids = append(ids, p.ID)
		}

		posts, err := l.Posts(ctx, 10, ids[1])
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(posts) != 1 {
			t.Fatalf("expected 1, got %d", len(posts))
		}
	})

	t.Run("first <= 0 returns empty", func(t *testing.T) {
		l := New()
		if _, err := l.NewPost(ctx, post.Post{Title: "T"}); err != nil {
			t.Fatal(err)
		}
		posts, err := l.Posts(ctx, 0, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(posts) != 0 {
			t.Fatalf("expected empty, got %d", len(posts))
		}
	})
}

func TestSetCommentsEnabled(t *testing.T) {
	ctx := context.Background()

	t.Run("toggles flag", func(t *testing.T) {
		l := New()
		p, _ := l.NewPost(ctx, post.Post{Title: "T", CommentsEnabled: true})

		if err := l.SetCommentsEnabled(ctx, p.ID, false); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, _ := l.Post(ctx, p.ID)
		if got.CommentsEnabled {
			t.Fatal("expected comments disabled")
		}
	})

	t.Run("not found", func(t *testing.T) {
		l := New()
		err := l.SetCommentsEnabled(ctx, 999, false)
		if !errors.Is(err, errs.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}
