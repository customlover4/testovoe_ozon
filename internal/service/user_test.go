package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"testovoe/internal/entities/user"
	"testovoe/pkg/errs"
	"testovoe/pkg/password"
	"testovoe/pkg/token"
)

var testSecret = []byte("super-secret-key-for-tests-123456")

func newAuthTestService(storage *mockStorage) *service {
	return &service{
		stor:   storage,
		secret: testSecret,
	}
}

func TestLoginPasswordValidation(t *testing.T) {
	svc := newAuthTestService(&mockStorage{})

	t.Run("empty login", func(t *testing.T) {
		err := svc.LoginPasswordValidation("", "password")
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("short login", func(t *testing.T) {
		err := svc.LoginPasswordValidation("a", "password")
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("empty password", func(t *testing.T) {
		err := svc.LoginPasswordValidation("login", "")
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("short password", func(t *testing.T) {
		err := svc.LoginPasswordValidation("login", "a")
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("valid", func(t *testing.T) {
		err := svc.LoginPasswordValidation("login", "password")
		assert.NoError(t, err)
	})
}

func TestRegistration(t *testing.T) {
	ctx := context.Background()

	t.Run("validation error", func(t *testing.T) {
		svc := newAuthTestService(&mockStorage{})
		_, err := svc.Registration(ctx, "a", "password")
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("login not unique", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newAuthTestService(storage)

		storage.On("NewUser", ctx, "alice", mock.AnythingOfType("string")).Return(uint64(0), errs.ErrNotUnique)

		_, err := svc.Registration(ctx, "alice", "password")
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("storage error", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newAuthTestService(storage)

		storage.On("NewUser", ctx, "alice", mock.AnythingOfType("string")).Return(uint64(0), errors.New("db error"))

		_, err := svc.Registration(ctx, "alice", "password")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "internal.service.Registration")
	})

	t.Run("success", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newAuthTestService(storage)

		var capturedHash string
		storage.On("NewUser", ctx, "alice", mock.AnythingOfType("string")).
			Run(func(args mock.Arguments) {
				capturedHash = args.String(2)
			}).
			Return(uint64(42), nil)

		tokenStr, err := svc.Registration(ctx, "alice", "password")
		require.NoError(t, err)
		assert.NotEmpty(t, tokenStr)

		assert.NotEqual(t, "password", capturedHash)
		assert.NotEmpty(t, capturedHash)

		subject, err := token.Parse(tokenStr, testSecret)
		require.NoError(t, err)
		assert.Equal(t, "42", subject)
	})
}

func TestLogin(t *testing.T) {
	ctx := context.Background()

	t.Run("validation error", func(t *testing.T) {
		svc := newAuthTestService(&mockStorage{})
		_, err := svc.Login(ctx, "", "password")
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("user not found", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newAuthTestService(storage)

		storage.On("User", ctx, "unknown").Return(user.User{}, errs.ErrNotFound)

		_, err := svc.Login(ctx, "unknown", "password")
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("wrong password", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newAuthTestService(storage)

		correctHash, _ := password.Generate("correct_password")

		storage.On("User", ctx, "alice").Return(user.User{
			ID: 42, Login: "alice", Password: correctHash,
		}, nil)

		_, err := svc.Login(ctx, "alice", "wrong_password")
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("success", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newAuthTestService(storage)

		correctHash, _ := password.Generate("correct_password")

		storage.On("User", ctx, "alice").Return(user.User{
			ID: 42, Login: "alice", Password: correctHash,
		}, nil)

		tokenStr, err := svc.Login(ctx, "alice", "correct_password")
		require.NoError(t, err)
		assert.NotEmpty(t, tokenStr)

		subject, err := token.Parse(tokenStr, testSecret)
		require.NoError(t, err)
		assert.Equal(t, "42", subject)
	})

	t.Run("storage error on user lookup", func(t *testing.T) {
		storage := new(mockStorage)
		svc := newAuthTestService(storage)

		storage.On("User", ctx, "alice").Return(user.User{}, errors.New("db error"))

		_, err := svc.Login(ctx, "alice", "password")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "internal.service.Login")
	})
}
