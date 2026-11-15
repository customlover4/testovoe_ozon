package local

import (
	"context"
	"testovoe/internal/entities/user"
	"testovoe/pkg/errs"
)

func (l *Local) NewUser(ctx context.Context, login, password string) (uint64, error) {
	l.usersMu.Lock()
	defer l.usersMu.Unlock()

	if _, ok := l.usersByName[login]; ok {
		return 0, errs.ErrNotUnique
	}

	newId := l.NewUserID()

	l.users[newId] = &user.User{
		ID:       newId,
		Login:    login,
		Password: password,
	}
	l.usersByName[login] = newId

	return newId, nil
}

func (l *Local) User(ctx context.Context, login string) (user.User, error) {
	l.usersMu.RLock()
	defer l.usersMu.RUnlock()

	if v, ok := l.usersByName[login]; !ok {
		return user.User{}, errs.ErrNotFound
	} else {
		return *l.users[v], nil
	}
}
