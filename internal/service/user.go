package service

import (
	"errors"
	"fmt"
	"strconv"
	"testovoe/internal/entities/user"
	"testovoe/pkg/errs"
	"testovoe/pkg/password"
	"testovoe/pkg/token"
	"time"
)

type auth interface {
	NewUser(login, password string) (uint64, error)
	User(login string) (user.User, error)
}

func (s *service) LoginPasswordValidation(login, pswd string) error {
	if login == "" || len(login) < 2 {
		return errs.NewValidation("login", "empty field or too short")
	}
	if pswd == "" || len(pswd) < 2 {
		return errs.NewValidation("password", "empty field or too short")
	}

	return nil
}

func (s *service) Registration(login, pswd string) (string, error) {
	const op = "internal.service.Registration"

	if err := s.LoginPasswordValidation(login, pswd); err != nil {
		return "", err
	}

	hashed, err := password.Generate(pswd)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	id, err := s.stor.NewUser(login, hashed)
	if errors.Is(err, errs.ErrNotUnique) {
		return "", errs.NewValidation("login", "not unique")
	} else if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	t, err := token.GenerateToken(
		strconv.FormatUint(id, 10), s.secret, time.Hour*24,
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return t, nil
}

func (s *service) Login(login, pswd string) (string, error) {
	const op = "internal.service.Login"

	if err := s.LoginPasswordValidation(login, pswd); err != nil {
		return "", err
	}

	user, err := s.stor.User(login)
	if errors.Is(err, errs.ErrNotFound) {
		return "", errs.NewValidation("password", "wrong password")
	} else if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	same, err := password.Compare(pswd, user.Password)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if !same {
		return "", errs.NewValidation("password", "wrong password")
	}

	t, err := token.GenerateToken(
		strconv.FormatUint(user.ID, 10), s.secret, time.Hour*24,
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return t, nil
}
