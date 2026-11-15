package service

import (
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

func (s *service) LoginPasswordValidation(login, pswd string) (bool, error) {
	if login == "" || len(login) < 2 {
		return false, errs.NewValidation("login", "empty field or too short")
	}
	if pswd == "" || len(pswd) < 2 {
		return false, errs.NewValidation("password", "empty field or too short")
	}

	return true, nil
}

func (s *service) Registration(login, pswd string) (string, error) {
	const op = "internal.service.Registration"

	if valid, err := s.LoginPasswordValidation(login, pswd); !valid {
		return "", err
	}

	hashed, err := password.Generate(pswd)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	id, err := s.stor.NewUser(login, hashed)
	if err != nil {
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

	if valid, err := s.LoginPasswordValidation(login, pswd); !valid {
		return "", err
	}

	user, err := s.stor.User(login)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if user.Empty() {
		return "", nil // TODO:
	}

	same, err := password.Compare(pswd, user.Password)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if !same {
		return "", nil // TODO:
	}

	t, err := token.GenerateToken(
		strconv.FormatUint(user.ID, 10), s.secret, time.Hour*24,
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return t, nil
}
