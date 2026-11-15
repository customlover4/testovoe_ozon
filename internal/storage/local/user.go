package local

import (
	"strconv"
	"testovoe/internal/entities/user"
	"testovoe/pkg/errs"
)

func (l *Local) NewUser(login, password string) (uint64, error) {
	if _, ok := l.usersByName[login]; ok {
		return 0, errs.ErrNotUnique
	}
	
	l.lastUserID++

	id := strconv.FormatUint(l.lastUserID, 10)
	l.users[id] = &user.User{
		ID:       l.lastUserID,
		Login:    login,
		Password: password,
	}
	l.usersByName[login] = id

	return l.lastUserID, nil
}

func (l *Local) User(login string) (user.User, error) {
	if v, ok := l.usersByName[login]; !ok {
		return user.User{}, errs.ErrNotFound
	} else {
		return *l.users[v], nil
	}
}
