package local

import (
	"context"
	"testovoe/internal/entities/comment"
	"testovoe/pkg/errs"
)

func (l *Local) NewComment(ctx context.Context, c comment.Comment) (comment.Comment, error) {
	l.commentsMu.Lock()
	defer l.commentsMu.Unlock()

	if c.ParentID != 0 {
		parent, ok := l.comments[c.ParentID]
		if !ok || parent.PostID != c.PostID {
			return comment.Comment{}, errs.ErrNotFound
		}
	}

	c.ID = l.newCommentID()

	l.comments[c.ID] = &c

	key := commentThread{postID: c.PostID, parentID: c.ParentID}
	l.threads[key] = append(l.threads[key], c.ID)

	return c, nil
}

func (l *Local) Comment(ctx context.Context, id uint64) (comment.Comment, error) {
	l.commentsMu.RLock()
	defer l.commentsMu.RUnlock()

	c, ok := l.comments[id]
	if !ok {
		return comment.Comment{}, errs.ErrNotFound
	}
	return *c, nil
}

func (l *Local) Comments(ctx context.Context, postID, parentID uint64, first int, after uint64) ([]comment.Comment, error) {
	l.commentsMu.RLock()
	defer l.commentsMu.RUnlock()

	if first <= 0 {
		return []comment.Comment{}, nil
	}

	ids := l.threads[commentThread{postID: postID, parentID: parentID}]

	start := 0
	if after != 0 {
		start = -1
		for i, id := range ids {
			if id == after {
				start = i + 1
				break
			}
		}
		if start < 0 {
			return nil, errs.ErrNotFound
		}
	}

	end := start + first
	if end > len(ids) {
		end = len(ids)
	}

	out := make([]comment.Comment, 0, end-start)
	for _, id := range ids[start:end] {
		out = append(out, *l.comments[id])
	}
	return out, nil
}

func (l *Local) CommentTree(ctx context.Context, postID uint64, first int, after uint64) ([]comment.Comment, uint64, error) {
	l.commentsMu.RLock()
	defer l.commentsMu.RUnlock()

	rootIDs := l.threads[commentThread{postID: postID}]

	if first <= 0 {
		return []comment.Comment{}, 0, nil
	}

	start := 0
	if after != 0 {
		start = -1
		for i, id := range rootIDs {
			if id == after {
				start = i + 1
				break
			}
		}
		if start < 0 {
			return nil, 0, errs.ErrNotFound
		}
	}

	end := start + first
	if end > len(rootIDs) {
		end = len(rootIDs)
	}

	var next uint64
	if start < end && end < len(rootIDs) {
		next = rootIDs[end-1]
	}

	stack := make([]uint64, 0, end-start)
	for i := end - 1; i >= start; i-- {
		stack = append(stack, rootIDs[i])
	}

	comments := make([]comment.Comment, 0, end-start)
	for len(stack) > 0 {
		last := len(stack) - 1
		id := stack[last]
		stack = stack[:last]

		c := l.comments[id]
		comments = append(comments, *c)

		childIDs := l.threads[commentThread{postID: postID, parentID: id}]
		for i := len(childIDs) - 1; i >= 0; i-- {
			stack = append(stack, childIDs[i])
		}
	}

	return comments, next, nil
}
