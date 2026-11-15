package storage

import "testovoe/internal/entities/user"

type userMethods interface {
	NewUser(login, password string) (uint64, error)
	User(login string) (user.User, error)
}

type postMethods interface {
	NewPost()
	Post()
	Posts()
	DeletePost()
}

type contract interface {
	userMethods
	// post
}

type storage struct {
	s contract
}

func New(instance contract) *storage {
	return &storage{
		s: instance,
	}
}
