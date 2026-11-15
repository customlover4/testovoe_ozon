package graph

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testovoe/pkg/errs"
	"testovoe/pkg/token"
)

func TestAuthenticate(t *testing.T) {
	secret := []byte("test-secret")

	runAuth := func(t *testing.T, req *http.Request) (userID uint64, ok bool) {
		t.Helper()

		var capturedCtx context.Context
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedCtx = r.Context()
		})

		authenticate(secret, next).ServeHTTP(httptest.NewRecorder(), req)
		require.NotNil(t, capturedCtx)

		userID, ok = capturedCtx.Value(userIDContextKey{}).(uint64)
		return
	}

	t.Run("no token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/query", nil)
		_, ok := runAuth(t, req)
		assert.False(t, ok)
	})

	t.Run("valid bearer token", func(t *testing.T) {
		tok, err := token.GenerateToken("42", secret, time.Hour)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/query", nil)
		req.Header.Set("Authorization", "Bearer "+tok)

		id, ok := runAuth(t, req)
		assert.True(t, ok)
		assert.Equal(t, uint64(42), id)
	})

	t.Run("valid cookie", func(t *testing.T) {
		tok, err := token.GenerateToken("7", secret, time.Hour)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/query", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: tok})

		id, ok := runAuth(t, req)
		assert.True(t, ok)
		assert.Equal(t, uint64(7), id)
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/query", nil)
		req.Header.Set("Authorization", "Bearer garbage")
		_, ok := runAuth(t, req)
		assert.False(t, ok)
	})

	t.Run("token with zero id", func(t *testing.T) {
		tok, err := token.GenerateToken("0", secret, time.Hour)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/query", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		_, ok := runAuth(t, req)
		assert.False(t, ok)
	})

	t.Run("bearer takes precedence over cookie", func(t *testing.T) {
		bearerTok, _ := token.GenerateToken("42", secret, time.Hour)
		cookieTok, _ := token.GenerateToken("7", secret, time.Hour)

		req := httptest.NewRequest(http.MethodPost, "/query", nil)
		req.Header.Set("Authorization", "Bearer "+bearerTok)
		req.AddCookie(&http.Cookie{Name: "token", Value: cookieTok})

		id, ok := runAuth(t, req)
		assert.True(t, ok)
		assert.Equal(t, uint64(42), id)
	})
}

func TestBearerToken(t *testing.T) {
	cases := []struct {
		name   string
		header string
		want   string
	}{
		{"empty", "", ""},
		{"no prefix", "abc", ""},
		{"basic", "Basic abc", ""},
		{"bearer", "Bearer abc", "abc"},
		{"bearer with spaces", "Bearer   abc  ", "abc"},
		{"case sensitive", "bearer abc", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if c.header != "" {
				req.Header.Set("Authorization", c.header)
			}
			assert.Equal(t, c.want, bearerToken(req))
		})
	}
}

func TestErrorPresenter(t *testing.T) {
	log := zerolog.Nop()
	present := errorPresenter(log)
	ctx := context.Background()

	t.Run("validation", func(t *testing.T) {
		err := errs.NewValidation("login", "too short")
		res := present(ctx, err)
		assert.Equal(t, "VALIDATION", res.Extensions["code"])
	})

	t.Run("not found", func(t *testing.T) {
		res := present(ctx, errs.ErrNotFound)
		assert.Equal(t, "not found", res.Message)
		assert.Equal(t, "NOT_FOUND", res.Extensions["code"])
	})

	t.Run("forbidden", func(t *testing.T) {
		res := present(ctx, errs.ErrForbidden)
		assert.Equal(t, "forbidden", res.Message)
		assert.Equal(t, "FORBIDDEN", res.Extensions["code"])
	})

	t.Run("internal (unknown)", func(t *testing.T) {
		res := present(ctx, errors.New("boom"))
		assert.Equal(t, "internal server error", res.Message)
		assert.Equal(t, "INTERNAL", res.Extensions["code"])
	})

	t.Run("wrapped error is unwrapped", func(t *testing.T) {
		wrapped := errors.Join(errs.ErrForbidden, errors.New("boom"))
		res := present(ctx, wrapped)
		assert.Equal(t, "FORBIDDEN", res.Extensions["code"])
	})
}
