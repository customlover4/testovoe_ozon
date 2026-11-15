package local

import (
	"sync"
	"testovoe/internal/entities/comment"
	"testovoe/internal/entities/post"
	"testovoe/internal/entities/user"
)

type commentThread struct {
	postID   uint64
	parentID uint64
}

type Local struct {
	usersMu    *sync.RWMutex
	postsMu    *sync.RWMutex
	commentsMu *sync.RWMutex

	users       map[uint64]*user.User
	usersByName map[string]uint64
	lastUserID  uint64

	posts      map[uint64]*post.Post
	postsID    []uint64
	lastPostID uint64

	comments      map[uint64]*comment.Comment
	threads       map[commentThread][]uint64
	lastCommentID uint64
}

func New() *Local {
	return &Local{
		usersMu:     &sync.RWMutex{},
		users:       make(map[uint64]*user.User),
		usersByName: make(map[string]uint64),
		lastUserID:  0,

		postsMu:    &sync.RWMutex{},
		posts:      make(map[uint64]*post.Post),
		postsID:    make([]uint64, 0),
		lastPostID: 0,

		commentsMu:    &sync.RWMutex{},
		comments:      make(map[uint64]*comment.Comment),
		threads:       make(map[commentThread][]uint64),
		lastCommentID: 0,
	}
}

func (l *Local) newUserID() uint64 {
	l.lastUserID++
	return l.lastUserID
}

func (l *Local) newPostID() uint64 {
	l.lastPostID++
	return l.lastPostID
}

func (l *Local) newCommentID() uint64 {
	l.lastCommentID++
	return l.lastCommentID
}
