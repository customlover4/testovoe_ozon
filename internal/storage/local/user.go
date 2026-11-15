package local

import (
	"testovoe/internal/entities/user"
)

func (l *Local) NewUser(login, password string) (uint64, error) {
	return 0, nil
}

func (l *Local) User(login string) (user.User, error) {
	return user.User{}, nil
}
