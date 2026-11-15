package local

import (
	"sync"
	"testovoe/internal/entities/user"
)

type Local struct {
	mu *sync.RWMutex

	users       map[string]*user.User
	usersByName map[string]string
	lastUserID  uint64
}

func New() *Local {
	return &Local{
		mu:          &sync.RWMutex{},
		users:       make(map[string]*user.User),
		usersByName: make(map[string]string),
		lastUserID:  0,
	}
}
