package web

import (
	"net/http"
	"testovoe/internal/handlers"

	"github.com/gin-gonic/gin"
)

func NewRouter(debug bool) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	g := gin.New()
	SetupVendor(g)
	handlers.SetupRoutes(g)

	return g
}
