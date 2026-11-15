package local

import (
	"context"
	"testovoe/internal/entities/post"
	"testovoe/pkg/errs"
)

func (l *Local) NewPost(ctx context.Context, p post.Post) (post.Post, error) {
	l.postsMu.Lock()
	defer l.postsMu.Unlock()

	p.ID = l.newPostID()
	l.posts[p.ID] = &p
	l.postsID = append(l.postsID, p.ID)
	return p, nil
}

func (l *Local) Post(ctx context.Context, id uint64) (post.Post, error) {
	l.postsMu.RLock()
	defer l.postsMu.RUnlock()

	p, ok := l.posts[id]
	if !ok {
		return post.Post{}, errs.ErrNotFound
	}
	return *p, nil
}

func (l *Local) Posts(ctx context.Context, first int, after uint64) ([]post.Post, error) {
	l.postsMu.RLock()
	defer l.postsMu.RUnlock()

	if first <= 0 {
		return []post.Post{}, nil
	}

	start := 0
	if after != 0 {
		for i, id := range l.postsID {
			if id == after {
				start = i + 1
				break
			}
		}
	}

	end := start + first
	if end > len(l.postsID) {
		end = len(l.postsID)
	}

	out := make([]post.Post, 0, end-start)
	for _, id := range l.postsID[start:end] {
		out = append(out, *l.posts[id])
	}
	return out, nil
}

func (l *Local) SetCommentsEnabled(ctx context.Context, postID uint64, enabled bool) error {
	l.postsMu.Lock()
	defer l.postsMu.Unlock()

	p, ok := l.posts[postID]
	if !ok {
		return errs.ErrNotFound
	}
	p.CommentsEnabled = enabled
	return nil
}
