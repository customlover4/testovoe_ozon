package middleware

import (
	"net/http"
	"testovoe/pkg/token"

	"github.com/gin-gonic/gin"
)

func Auth(secret []byte) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		t, err := ctx.Cookie("token")
		if err != nil {
			ctx.Redirect(http.StatusSeeOther, "/login")
			ctx.Abort()
			return
		}

		id, err := token.Parse(t, secret)
		if err != nil {
			ctx.SetCookie("token", "", -1, "/", "", false, true)
			ctx.Redirect(http.StatusSeeOther, "/login")
			ctx.Abort()
			return
		}

		ctx.Set("ID", id)
		ctx.Next()
	}
}

func OptionalAuth(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		t, err := c.Cookie("token")
		if err == nil {
			if id, err := token.Parse(t, secret); err == nil {
				c.Set("ID", id)
			}
		}
		c.Next()
	}
}

func NotAuth(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		t, err := c.Cookie("token")
		if err == nil {
			if _, err := token.Parse(t, secret); err == nil {
				c.Redirect(http.StatusSeeOther, "/")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
