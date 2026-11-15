package graph

import (
	"context"
	"errors"
	"strconv"

	"testovoe/internal/entities/comment"
	"testovoe/internal/entities/post"
	"testovoe/internal/graphql/model"
	"testovoe/pkg/errs"
)

func (r *mutationResolver) Register(ctx context.Context, login string, password string) (*model.AuthPayload, error) {
	token, err := r.service.Registration(ctx, login, password)
	if err != nil {
		return nil, err
	}
	return &model.AuthPayload{Token: token}, nil
}

func (r *mutationResolver) Login(ctx context.Context, login string, password string) (*model.AuthPayload, error) {
	token, err := r.service.Login(ctx, login, password)
	if err != nil {
		return nil, err
	}
	return &model.AuthPayload{Token: token}, nil
}

func (r *mutationResolver) CreatePost(ctx context.Context, input model.CreatePostInput) (*model.Post, error) {
	authorID, err := authenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	commentsEnabled := true
	if input.CommentsEnabled != nil {
		commentsEnabled = *input.CommentsEnabled
	}
	created, err := r.service.NewPost(ctx, post.Post{
		Title: input.Title, Body: input.Body, AuthorID: authorID, CommentsEnabled: commentsEnabled,
	})
	if err != nil {
		return nil, err
	}
	return postModel(created), nil
}

func (r *mutationResolver) SetPostCommentsEnabled(ctx context.Context, postID string, enabled bool) (*model.Post, error) {
	authorID, err := authenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseID(postID, "post_id")
	if err != nil {
		return nil, err
	}
	updated, err := r.service.SetCommentsEnabled(ctx, id, authorID, enabled)
	if err != nil {
		return nil, err
	}
	return postModel(updated), nil
}

func (r *mutationResolver) CreateComment(ctx context.Context, input model.CreateCommentInput) (*model.Comment, error) {
	authorID, err := authenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	postID, err := parseID(input.PostID, "post_id")
	if err != nil {
		return nil, err
	}
	var parentID uint64
	if input.ParentID != nil {
		parentID, err = parseID(*input.ParentID, "parent_id")
		if err != nil {
			return nil, err
		}
	}
	created, err := r.service.NewComment(ctx, comment.Comment{
		PostID: postID, ParentID: parentID, AuthorID: authorID, Text: input.Text,
	})
	if err != nil {
		return nil, err
	}
	r.broker.publish(created)
	return commentModel(created), nil
}

func (r *postResolver) Comments(ctx context.Context, obj *model.Post, first *int, after *string) (*model.CommentTree, error) {
	postID, err := parseID(obj.ID, "post_id")
	if err != nil {
		return nil, err
	}
	comments, next, err := r.service.CommentTree(ctx, postID, pageSize(first), cursor(after))
	if err != nil {
		return nil, err
	}
	return commentTreeModel(comments, next), nil
}

func (r *queryResolver) Posts(ctx context.Context, first *int, after *string) (*model.PostConnection, error) {
	posts, next, err := r.service.Posts(ctx, pageSize(first), cursor(after))
	if err != nil {
		return nil, err
	}
	items := make([]*model.Post, 0, len(posts))
	for _, p := range posts {
		items = append(items, postModel(p))
	}
	return &model.PostConnection{Items: items, PageInfo: pageInfo(next)}, nil
}

func (r *queryResolver) Post(ctx context.Context, id string) (*model.Post, error) {
	postID, err := parseID(id, "id")
	if err != nil {
		return nil, err
	}
	p, err := r.service.Post(ctx, postID)
	if errors.Is(err, errs.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return postModel(p), nil
}

func (r *subscriptionResolver) CommentAdded(ctx context.Context, postID string) (<-chan *model.Comment, error) {
	id, err := parseID(postID, "post_id")
	if err != nil {
		return nil, err
	}
	source := r.broker.subscribe(ctx, id)
	output := make(chan *model.Comment, 32)
	go func() {
		defer close(output)
		for c := range source {
			select {
			case output <- commentModel(c):
			case <-ctx.Done():
				return
			}
		}
	}()
	return output, nil
}

type userIDContextKey struct{}

func WithUserID(ctx context.Context, id uint64) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, id)
}

func authenticatedUser(ctx context.Context) (uint64, error) {
	id, ok := ctx.Value(userIDContextKey{}).(uint64)
	if !ok || id == 0 {
		return 0, errs.ErrForbidden
	}
	return id, nil
}

func parseID(value, field string) (uint64, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, errs.NewValidation(field, "invalid value")
	}
	return id, nil
}

func cursor(value *string) uint64 {
	if value == nil {
		return 0
	}
	id, err := strconv.ParseUint(*value, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func pageSize(first *int) int {
	if first == nil {
		return 10
	}
	return *first
}

func postModel(p post.Post) *model.Post {
	return &model.Post{
		ID: strconv.FormatUint(p.ID, 10), Title: p.Title, Body: p.Body,
		AuthorID: strconv.FormatUint(p.AuthorID, 10), CommentsEnabled: p.CommentsEnabled,
	}
}

func commentModel(c comment.Comment) *model.Comment {
	result := &model.Comment{
		ID: strconv.FormatUint(c.ID, 10), PostID: strconv.FormatUint(c.PostID, 10),
		AuthorID: strconv.FormatUint(c.AuthorID, 10), Text: c.Text,
	}
	if c.ParentID != 0 {
		parentID := strconv.FormatUint(c.ParentID, 10)
		result.ParentID = &parentID
	}
	return result
}

func commentTreeModel(comments []comment.Comment, next uint64) *model.CommentTree {
	items := make([]*model.Comment, 0, len(comments))
	for _, c := range comments {
		items = append(items, commentModel(c))
	}
	return &model.CommentTree{Items: items, PageInfo: pageInfo(next)}
}

func pageInfo(next uint64) *model.PageInfo {
	info := &model.PageInfo{HasNextPage: next != 0}
	if next != 0 {
		cursor := strconv.FormatUint(next, 10)
		info.EndCursor = &cursor
	}
	return info
}

func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

func (r *Resolver) Post() PostResolver { return &postResolver{r} }

func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

type (
	mutationResolver     struct{ *Resolver }
	postResolver         struct{ *Resolver }
	queryResolver        struct{ *Resolver }
	subscriptionResolver struct{ *Resolver }
)
