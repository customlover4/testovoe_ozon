package graph

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/rs/zerolog"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"testovoe/pkg/errs"
	"testovoe/pkg/token"
)

func NewServer(service Service, secret []byte, log zerolog.Logger) http.Handler {
	server := handler.New(NewExecutableSchema(Config{Resolvers: NewResolver(service)}))
	server.AddTransport(transport.Options{})
	server.AddTransport(transport.GET{})
	server.AddTransport(transport.POST{})
	server.AddTransport(transport.Websocket{KeepAlivePingInterval: 15 * time.Second})

	server.Use(extension.Introspection{})

	server.SetErrorPresenter(errorPresenter(log))

	return authenticate(secret, server)
}

func authenticate(secret []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		value := bearerToken(request)
		if value == "" {
			if cookie, err := request.Cookie("token"); err == nil {
				value = cookie.Value
			}
		}

		if value != "" {
			if subject, err := token.Parse(value, secret); err == nil {
				if id, err := strconv.ParseUint(subject, 10, 64); err == nil && id != 0 {
					request = request.WithContext(WithUserID(request.Context(), id))
				}
			}
		}
		next.ServeHTTP(w, request)
	})
}

func bearerToken(request *http.Request) string {
	const prefix = "Bearer "
	value := request.Header.Get("Authorization")
	if strings.HasPrefix(value, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(value, prefix))
	}
	return ""
}

func errorPresenter(log zerolog.Logger) graphql.ErrorPresenterFunc {
	return func(ctx context.Context, err error) *gqlerror.Error {
		result := graphql.DefaultErrorPresenter(ctx, err)
		switch {
		case errors.Is(err, errs.ErrValidation):
			result.Extensions = map[string]any{"code": "VALIDATION"}
		case errors.Is(err, errs.ErrNotFound):
			result.Message = "not found"
			result.Extensions = map[string]any{"code": "NOT_FOUND"}
		case errors.Is(err, errs.ErrForbidden):
			result.Message = "forbidden"
			result.Extensions = map[string]any{"code": "FORBIDDEN"}
		default:
			log.Error().Err(err).Msg("graphql request failed")
			result.Message = "internal server error"
			result.Extensions = map[string]any{"code": "INTERNAL"}
		}
		return result
	}
}
