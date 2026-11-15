package graph

import (
	"context"
	"sync"

	"testovoe/internal/entities/comment"
	"testovoe/internal/entities/post"
)

type Service interface {
	Registration(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, error)
	NewPost(ctx context.Context, p post.Post) (post.Post, error)
	Posts(ctx context.Context, first int, after uint64) ([]post.Post, uint64, error)
	Post(ctx context.Context, id uint64) (post.Post, error)
	SetCommentsEnabled(ctx context.Context, postID, authorID uint64, enabled bool) (post.Post, error)
	NewComment(ctx context.Context, c comment.Comment) (comment.Comment, error)
	CommentTree(ctx context.Context, postID uint64, first int, after uint64) ([]comment.Comment, uint64, error)
}

type Resolver struct {
	service Service
	broker  *commentBroker
}

func NewResolver(service Service) *Resolver {
	return &Resolver{service: service, broker: newCommentBroker()}
}

type commentBroker struct {
	mu          sync.RWMutex
	subscribers map[uint64]map[chan comment.Comment]struct{}
}

func newCommentBroker() *commentBroker {
	return &commentBroker{subscribers: make(map[uint64]map[chan comment.Comment]struct{})}
}

func (b *commentBroker) subscribe(ctx context.Context, postID uint64) <-chan comment.Comment {
	ch := make(chan comment.Comment, 32)
	b.mu.Lock()
	if b.subscribers[postID] == nil {
		b.subscribers[postID] = make(map[chan comment.Comment]struct{})
	}
	b.subscribers[postID][ch] = struct{}{}
	b.mu.Unlock()

	go func() {
		<-ctx.Done()
		b.mu.Lock()
		delete(b.subscribers[postID], ch)
		if len(b.subscribers[postID]) == 0 {
			delete(b.subscribers, postID)
		}
		close(ch)
		b.mu.Unlock()
	}()
	return ch
}

func (b *commentBroker) publish(c comment.Comment) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subscribers[c.PostID] {
		select {
		case ch <- c:
		default:
		}
	}
}
