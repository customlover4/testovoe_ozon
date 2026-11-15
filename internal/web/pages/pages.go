package pages

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	MainHTML = "index.html"
)

func MainPage() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, exists := ctx.Get("ID")
		ctx.HTML(http.StatusOK, MainHTML, gin.H{
			"IsLoggedIn": exists,
		})
	}
}
