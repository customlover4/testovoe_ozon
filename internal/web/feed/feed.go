package feed

import (
	"context"
	"net/http"
	"strconv"
	"testovoe/internal/entities/post"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

var (
	FeedHTML       = "post.html"
	CreatePostHTML = "new_post.html"
)

type FeedService interface {
	NewPost(ctx context.Context, p post.Post) error
	Posts(ctx context.Context, first int, after uint64) ([]*post.Post, uint64, error)
}

func Feed(log zerolog.Logger, svc FeedService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var q struct {
			First int    `form:"first"`
			After uint64 `form:"after"`
		}
		_ = ctx.ShouldBindQuery(&q)

		posts, next, err := svc.Posts(ctx.Request.Context(), q.First, q.After)
		if err != nil {
			log.Error().Err(err).Msg("error on feed loading")
			ctx.HTML(http.StatusInternalServerError, FeedHTML, gin.H{
				"error": "feed failed",
			})
			return
		}

		ctx.HTML(http.StatusOK, FeedHTML, gin.H{
			"Posts": posts,
			"Next":  next,
		})
	}
}

func CreatePostPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, CreatePostHTML, nil)
}

type NewPost struct {
	Title           string `form:"title" binding:"required"`
	Body            string `form:"body" binding:"required"`
	CommentsEnabled string `form:"comments-enabled"`
}

func CreatePost(log zerolog.Logger, svc FeedService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var f NewPost

		if err := ctx.ShouldBind(&f); err != nil {
			ctx.HTML(http.StatusBadRequest, CreatePostHTML, gin.H{
				"error": "something went wrong",
			})
			return
		}

		authorID, exists := ctx.Get("ID")
		if !exists {
			ctx.Redirect(http.StatusSeeOther, "/login")
			return
		}

		sID, ok := authorID.(string)
		if !ok {
			log.Error().Msg("wrong data in user auth context")
			ctx.HTML(http.StatusInternalServerError, CreatePostHTML, gin.H{
				"error": "internal error",
			})
			return
		}

		id, err := strconv.ParseUint(sID, 10, 64)
		if err != nil {
			log.Error().Msg("wrong data in user auth context")
			ctx.HTML(http.StatusInternalServerError, CreatePostHTML, gin.H{
				"error": "internal error",
			})
			return
		}

		err = svc.NewPost(ctx.Request.Context(), post.Post{
			Title:           f.Title,
			Body:            f.Body,
			CommentsEnabled: f.CommentsEnabled == "true",
			AuthorID:        id,
		})
		if err != nil {
			log.Error().Err(err).Msg("error on creating post")
			ctx.HTML(http.StatusInternalServerError, CreatePostHTML, gin.H{
				"error": "internal error",
			})
			return
		}

		ctx.Redirect(http.StatusSeeOther, "/feed")
	}
}
