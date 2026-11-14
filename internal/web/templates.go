package web

import "github.com/gin-gonic/gin"

func SetupVendor(g *gin.Engine) {
	g.Static("/static", "./static")

	g.Delims("[[", "]]")
	g.LoadHTMLGlob("./templates/**/*.html")
}
