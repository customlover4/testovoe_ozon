package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(debug bool) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	g := gin.New()

	g.GET("/", func(ctx *gin.Context) {
		ctx.String(200, "string", "sldgjldfj")
	})

	return g
}
