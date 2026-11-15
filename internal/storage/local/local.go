package local

import "testovoe/internal/entities/user"

type Local struct {
	users       map[string]*user.User
	usersByName map[string]string
	lastUserID  uint64
}

func New() *Local {
	return &Local{
		users:       make(map[string]*user.User),
		usersByName: make(map[string]string),
		lastUserID:  0,
	}
}
