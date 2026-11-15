package web

import (
	"net/http"
	"testovoe/internal/web/auth"
	"testovoe/internal/web/middleware"
	"testovoe/internal/web/pages"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type svc interface {
	auth.UserService
}

func NewRouter(debug bool, log zerolog.Logger, s svc, secret []byte) http.Handler {
	if debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	g := gin.New()
	SetupVendor(g)

	g.Use(gin.Recovery())

	optional := g.Group("/", middleware.OptionalAuth(secret))
	{
		optional.GET("/", pages.MainPage())
	}

	guest := g.Group("/", middleware.NotAuth(secret))
	{
		guest.GET("/registration", auth.RegistrationPage)
		guest.POST("/registration", auth.Registration(log, s))

		guest.GET("/login", auth.LoginPage)
		guest.POST("/login", auth.Login(log, s))
	}

	g.POST("/logout", auth.Logout(log, s))

	return g
}
