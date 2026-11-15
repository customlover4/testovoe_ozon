package postgres

import (
	"context"
	"errors"
	"fmt"
	"testovoe/internal/entities/user"
	"testovoe/pkg/errs"

	"github.com/jackc/pgx/v5"
)

func (s *Store) NewUser(ctx context.Context, login, password string) (uint64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id`, login, password).Scan(&id)
	if isUniqueViolation(err) {
		return 0, errs.ErrNotUnique
	}
	if err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}
	return uint64(id), nil
}

func (s *Store) User(ctx context.Context, login string) (user.User, error) {
	var u user.User
	var id int64
	err := s.pool.QueryRow(ctx, `SELECT id, login, password FROM users WHERE login = $1`, login).Scan(&id, &u.Login, &u.Password)
	if errors.Is(err, pgx.ErrNoRows) {
		return user.User{}, errs.ErrNotFound
	}
	if err != nil {
		return user.User{}, fmt.Errorf("select user: %w", err)
	}
	u.ID = uint64(id)
	return u, nil
}
