package local

import (
	"sync"
	"testovoe/internal/entities/post"
	"testovoe/internal/entities/user"
)

type Local struct {
	usersMu *sync.RWMutex
	postsMu *sync.RWMutex

	users       map[uint64]*user.User
	usersByName map[string]uint64
	lastUserID  uint64

	posts      map[uint64]*post.Post
	postsID    []uint64
	lastPostID uint64
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
	}
}

func (l *Local) NewUserID() uint64 {
	l.lastUserID++
	return l.lastUserID
}

func (l *Local) NewPostID() uint64 {
	l.lastPostID++
	return l.lastPostID
}
