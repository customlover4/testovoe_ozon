package pages

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	MainHTML = "index.html"
)

func MainPage() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, exists := ctx.Get("ID")
		fmt.Println(exists)
		ctx.HTML(http.StatusOK, MainHTML, gin.H{
			"IsLoggedIn": exists,
		})
	}
}
