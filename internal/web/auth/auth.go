package auth

import (
	"context"
	"errors"
	"net/http"
	"testovoe/pkg/errs"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

const (
	RegistrationHTML = "reg.html"
	LoginHTML        = "log.html"
)

type UserService interface {
	Registration(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, error)
}

type UserForm struct {
	Login    string `form:"login" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func RegistrationPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, RegistrationHTML, nil)
}

func Registration(log zerolog.Logger, svc UserService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var form UserForm

		if err := ctx.ShouldBind(&form); err != nil {
			ctx.HTML(http.StatusBadRequest, RegistrationHTML, gin.H{
				"error": "something went wrong",
			})
			return
		}

		token, err := svc.Registration(ctx.Request.Context(), form.Login, form.Password)
		if errors.Is(err, errs.ErrValidation) {
			ctx.HTML(http.StatusUnprocessableEntity, RegistrationHTML, gin.H{
				"error": err.Error(),
			})
			return
		} else if err != nil {
			log.Error().Err(err).Msg("registration failed")

			ctx.HTML(http.StatusInternalServerError, RegistrationHTML, gin.H{
				"error": "internal server error",
			})
			return
		}

		ctx.SetCookie(
			"token", token, 86400, "/", "", false, true,
		)

		ctx.Redirect(http.StatusSeeOther, "/")
	}
}

func LoginPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, LoginHTML, nil)
}

func Login(log zerolog.Logger, svc UserService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var form UserForm
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.HTML(http.StatusBadRequest, RegistrationHTML, gin.H{
				"error": "something went wrong",
			})
			return
		}

		token, err := svc.Login(ctx.Request.Context(), form.Login, form.Password)
		if errors.Is(err, errs.ErrValidation) {
			ctx.HTML(http.StatusUnprocessableEntity, LoginHTML, gin.H{
				"error": err.Error(),
			})
			return
		} else if err != nil {
			log.Error().Err(err).Msg("registration failed")

			ctx.HTML(http.StatusInternalServerError, LoginHTML, gin.H{
				"error": "internal server error",
			})
			return
		}

		ctx.SetCookie(
			"token", token, 86400, "/", "", false, true,
		)

		ctx.Redirect(http.StatusSeeOther, "/")
	}
}

func Logout(log zerolog.Logger, svc UserService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.SetCookie("token", "", -1, "/", "", false, true)
		ctx.Redirect(http.StatusSeeOther, "/")
	}
}
