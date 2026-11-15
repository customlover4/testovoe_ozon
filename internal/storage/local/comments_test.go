package local

import (
	"context"
	"errors"
	"testing"

	"testovoe/internal/entities/comment"
	"testovoe/pkg/errs"
)

func TestNewComment(t *testing.T) {
	ctx := context.Background()

	t.Run("creates root comment", func(t *testing.T) {
		l := New()
		c, err := l.NewComment(ctx, comment.Comment{PostID: 1, ParentID: 0, Text: "root"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ID == 0 {
			t.Fatal("expected non-zero id")
		}

		got, err := l.Comment(ctx, c.ID)
		if err != nil {
			t.Fatalf("Comment: %v", err)
		}
		if got.Text != "root" || got.ParentID != 0 || got.PostID != 1 {
			t.Fatalf("unexpected comment: %+v", got)
		}
	})

	t.Run("creates reply", func(t *testing.T) {
		l := New()
		root, _ := l.NewComment(ctx, comment.Comment{PostID: 1, Text: "root"})
		reply, err := l.NewComment(ctx, comment.Comment{PostID: 1, ParentID: root.ID, Text: "reply"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if reply.ParentID != root.ID {
			t.Fatalf("expected parent %d, got %d", root.ID, reply.ParentID)
		}
	})

	t.Run("parent not found", func(t *testing.T) {
		l := New()
		_, err := l.NewComment(ctx, comment.Comment{PostID: 1, ParentID: 999, Text: "x"})
		if !errors.Is(err, errs.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("parent from another post", func(t *testing.T) {
		l := New()
		root, _ := l.NewComment(ctx, comment.Comment{PostID: 1, Text: "root"})
		_, err := l.NewComment(ctx, comment.Comment{PostID: 2, ParentID: root.ID, Text: "x"})
		if !errors.Is(err, errs.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("ids are unique", func(t *testing.T) {
		l := New()
		c1, _ := l.NewComment(ctx, comment.Comment{PostID: 1, Text: "a"})
		c2, _ := l.NewComment(ctx, comment.Comment{PostID: 1, Text: "b"})
		if c1.ID == c2.ID {
			t.Fatalf("expected unique ids, got %d", c1.ID)
		}
	})
}

func TestComment(t *testing.T) {
	ctx := context.Background()

	t.Run("not found", func(t *testing.T) {
		l := New()
		_, err := l.Comment(ctx, 999)
		if !errors.Is(err, errs.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("found", func(t *testing.T) {
		l := New()
		c, _ := l.NewComment(ctx, comment.Comment{PostID: 1, Text: "x"})
		got, err := l.Comment(ctx, c.ID)
		if err != nil {
			t.Fatalf("Comment: %v", err)
		}
		if got.ID != c.ID {
			t.Fatalf("expected id %d, got %d", c.ID, got.ID)
		}
	})
}

func TestComments(t *testing.T) {
	ctx := context.Background()

	t.Run("empty", func(t *testing.T) {
		l := New()
		got, err := l.Comments(ctx, 1, 0, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("expected empty, got %d", len(got))
		}
	})

	t.Run("first <= 0 returns empty", func(t *testing.T) {
		l := New()
		_, _ = l.NewComment(ctx, comment.Comment{PostID: 1, Text: "a"})
		got, err := l.Comments(ctx, 1, 0, 0, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("expected empty, got %d", len(got))
		}
	})

	t.Run("only direct children", func(t *testing.T) {
		l := New()
		root, _ := l.NewComment(ctx, comment.Comment{PostID: 1, Text: "root"})
		child1, _ := l.NewComment(ctx, comment.Comment{PostID: 1, ParentID: root.ID, Text: "c1"})
		_ = child1
		_, _ = l.NewComment(ctx, comment.Comment{PostID: 1, ParentID: root.ID, Text: "c2"})
		_, _ = l.NewComment(ctx, comment.Comment{PostID: 1, ParentID: child1.ID, Text: "grand"})

		got, err := l.Comments(ctx, 1, root.ID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("expected 2 direct children, got %d", len(got))
		}
	})

	t.Run("after cursor", func(t *testing.T) {
		l := New()
		var ids []uint64
		for i := 0; i < 3; i++ {
			c, _ := l.NewComment(ctx, comment.Comment{PostID: 1, Text: "x"})
			ids = append(ids, c.ID)
		}

		got, err := l.Comments(ctx, 1, 0, 10, ids[0])
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("expected 2, got %d", len(got))
		}
		if got[0].ID != ids[1] {
			t.Fatalf("expected first id %d, got %d", ids[1], got[0].ID)
		}
	})

	t.Run("after not found", func(t *testing.T) {
		l := New()
		_, _ = l.NewComment(ctx, comment.Comment{PostID: 1, Text: "x"})
		_, err := l.Comments(ctx, 1, 0, 10, 999)
		if !errors.Is(err, errs.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("first limits page", func(t *testing.T) {
		l := New()
		for i := 0; i < 5; i++ {
			_, _ = l.NewComment(ctx, comment.Comment{PostID: 1, Text: "x"})
		}
		got, err := l.Comments(ctx, 1, 0, 2, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("expected 2, got %d", len(got))
		}
	})
}

func TestCommentTree(t *testing.T) {
	ctx := context.Background()

	t.Run("empty", func(t *testing.T) {
		l := New()
		got, next, err := l.CommentTree(ctx, 1, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 || next != 0 {
			t.Fatalf("expected empty and next=0, got %d, %d", len(got), next)
		}
	})

	t.Run("first <= 0 returns empty", func(t *testing.T) {
		l := New()
		_, _ = l.NewComment(ctx, comment.Comment{PostID: 1, Text: "a"})
		got, next, err := l.CommentTree(ctx, 1, 0, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 || next != 0 {
			t.Fatalf("expected empty and next=0, got %d, %d", len(got), next)
		}
	})

	t.Run("preorder with descendants", func(t *testing.T) {
		l := New()
		root, _ := l.NewComment(ctx, comment.Comment{PostID: 1, Text: "root"})
		child1, _ := l.NewComment(ctx, comment.Comment{PostID: 1, ParentID: root.ID, Text: "c1"})
		child2, _ := l.NewComment(ctx, comment.Comment{PostID: 1, ParentID: root.ID, Text: "c2"})
		grand, _ := l.NewComment(ctx, comment.Comment{PostID: 1, ParentID: child1.ID, Text: "grand"})
		_ = child2

		got, next, err := l.CommentTree(ctx, 1, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if next != 0 {
			t.Fatalf("expected next=0, got %d", next)
		}
		want := []uint64{root.ID, child1.ID, grand.ID, child2.ID}
		if len(got) != len(want) {
			t.Fatalf("expected %d comments, got %d", len(want), len(got))
		}
		for i, id := range want {
			if got[i].ID != id {
				t.Fatalf("at %d: expected id %d, got %d", i, id, got[i].ID)
			}
		}
	})

	t.Run("next cursor when more roots", func(t *testing.T) {
		l := New()
		var roots []uint64
		for i := 0; i < 3; i++ {
			c, _ := l.NewComment(ctx, comment.Comment{PostID: 1, Text: "root"})
			roots = append(roots, c.ID)
		}

		got, next, err := l.CommentTree(ctx, 1, 2, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("expected 2, got %d", len(got))
		}
		if next != roots[1] {
			t.Fatalf("expected next %d, got %d", roots[1], next)
		}
	})

	t.Run("after root", func(t *testing.T) {
		l := New()
		var roots []uint64
		for i := 0; i < 3; i++ {
			c, _ := l.NewComment(ctx, comment.Comment{PostID: 1, Text: "root"})
			roots = append(roots, c.ID)
		}

		got, next, err := l.CommentTree(ctx, 1, 10, roots[0])
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("expected 2, got %d", len(got))
		}
		if got[0].ID != roots[1] {
			t.Fatalf("expected first id %d, got %d", roots[1], got[0].ID)
		}
		if next != 0 {
			t.Fatalf("expected next=0, got %d", next)
		}
	})

	t.Run("after not found", func(t *testing.T) {
		l := New()
		_, _ = l.NewComment(ctx, comment.Comment{PostID: 1, Text: "root"})
		_, _, err := l.CommentTree(ctx, 1, 10, 999)
		if !errors.Is(err, errs.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("isolated per post", func(t *testing.T) {
		l := New()
		c1, _ := l.NewComment(ctx, comment.Comment{PostID: 1, Text: "p1"})
		c2, _ := l.NewComment(ctx, comment.Comment{PostID: 2, Text: "p2"})

		got, _, err := l.CommentTree(ctx, 1, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || got[0].ID != c1.ID {
			t.Fatalf("expected only post 1 comment, got %+v", got)
		}

		got2, _, err := l.CommentTree(ctx, 2, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got2) != 1 || got2[0].ID != c2.ID {
			t.Fatalf("expected only post 2 comment, got %+v", got2)
		}
	})
}
