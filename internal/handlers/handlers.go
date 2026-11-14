package handlers

import (
	"net/http"
	"testovoe/internal/handlers/auth"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(g *gin.Engine) {
	g.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", nil)
	})

	auth.Setup(g)
}
