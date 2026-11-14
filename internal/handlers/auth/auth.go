package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	RegistrationHTML = "reg.html"
)

func Setup(g *gin.Engine) {
	g.GET("/registration")
	g.POST("/registration", func(ctx *gin.Context) {})
	g.POST("/login")
	g.POST("/logout")
}

func RegistrationPage() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, RegistrationHTML, nil)
	}
}

func Registration() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}
