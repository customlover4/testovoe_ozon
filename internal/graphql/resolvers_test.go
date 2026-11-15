package graph

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"testovoe/internal/entities/comment"
	"testovoe/internal/entities/post"
	"testovoe/internal/graphql/model"
	"testovoe/pkg/errs"
)

type mockService struct {
	mock.Mock
}

func (m *mockService) Registration(ctx context.Context, login, password string) (string, error) {
	args := m.Called(ctx, login, password)
	return args.String(0), args.Error(1)
}
func (m *mockService) Login(ctx context.Context, login, password string) (string, error) {
	args := m.Called(ctx, login, password)
	return args.String(0), args.Error(1)
}
func (m *mockService) NewPost(ctx context.Context, p post.Post) (post.Post, error) {
	args := m.Called(ctx, p)
	return args.Get(0).(post.Post), args.Error(1)
}
func (m *mockService) Posts(ctx context.Context, first int, after uint64) ([]post.Post, uint64, error) {
	args := m.Called(ctx, first, after)
	if args.Get(0) == nil {
		return nil, args.Get(1).(uint64), args.Error(2)
	}
	return args.Get(0).([]post.Post), args.Get(1).(uint64), args.Error(2)
}
func (m *mockService) Post(ctx context.Context, id uint64) (post.Post, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(post.Post), args.Error(1)
}
func (m *mockService) SetCommentsEnabled(ctx context.Context, postID, authorID uint64, enabled bool) (post.Post, error) {
	args := m.Called(ctx, postID, authorID, enabled)
	return args.Get(0).(post.Post), args.Error(1)
}
func (m *mockService) NewComment(ctx context.Context, c comment.Comment) (comment.Comment, error) {
	args := m.Called(ctx, c)
	return args.Get(0).(comment.Comment), args.Error(1)
}
func (m *mockService) CommentTree(ctx context.Context, postID uint64, first int, after uint64) ([]comment.Comment, uint64, error) {
	args := m.Called(ctx, postID, first, after)
	if args.Get(0) == nil {
		return nil, args.Get(1).(uint64), args.Error(2)
	}
	return args.Get(0).([]comment.Comment), args.Get(1).(uint64), args.Error(2)
}


func newTestResolver(svc Service) *Resolver {
	return &Resolver{service: svc, broker: newCommentBroker()}
}

func TestRegister(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		svc := new(mockService)
		r := &mutationResolver{newTestResolver(svc)}
		svc.On("Registration", ctx, "alice", "pass").Return("token123", nil)

		res, err := r.Register(ctx, "alice", "pass")
		require.NoError(t, err)
		assert.Equal(t, "token123", res.Token)
	})

	t.Run("error", func(t *testing.T) {
		svc := new(mockService)
		r := &mutationResolver{newTestResolver(svc)}
		svc.On("Registration", ctx, "alice", "pass").Return("", errs.NewValidation("login", "taken"))

		_, err := r.Register(ctx, "alice", "pass")
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})
}

func TestLogin(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		svc := new(mockService)
		r := &mutationResolver{newTestResolver(svc)}
		svc.On("Login", ctx, "alice", "pass").Return("token123", nil)

		res, err := r.Login(ctx, "alice", "pass")
		require.NoError(t, err)
		assert.Equal(t, "token123", res.Token)
	})

	t.Run("error", func(t *testing.T) {
		svc := new(mockService)
		r := &mutationResolver{newTestResolver(svc)}
		svc.On("Login", ctx, "alice", "pass").Return("", errs.NewValidation("password", "wrong"))

		_, err := r.Login(ctx, "alice", "pass")
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})
}

func TestCreatePost(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		svc := new(mockService)
		r := &mutationResolver{newTestResolver(svc)}

		_, err := r.CreatePost(context.Background(), model.CreatePostInput{Title: "T", Body: "B"})
		assert.True(t, errors.Is(err, errs.ErrForbidden))
	})

	t.Run("success with default commentsEnabled", func(t *testing.T) {
		svc := new(mockService)
		r := &mutationResolver{newTestResolver(svc)}
		ctx := WithUserID(context.Background(), 42)

		expected := post.Post{ID: 100, Title: "T", Body: "B", AuthorID: 42, CommentsEnabled: true}
		svc.On("NewPost", ctx, mock.MatchedBy(func(p post.Post) bool {
			return p.CommentsEnabled == true
		})).Return(expected, nil)

		res, err := r.CreatePost(ctx, model.CreatePostInput{Title: "T", Body: "B"})
		require.NoError(t, err)
		assert.Equal(t, "100", res.ID)
		assert.True(t, res.CommentsEnabled)
	})

	t.Run("success with commentsEnabled false", func(t *testing.T) {
		svc := new(mockService)
		r := &mutationResolver{newTestResolver(svc)}
		ctx := WithUserID(context.Background(), 42)
		f := false

		expected := post.Post{ID: 100, Title: "T", Body: "B", AuthorID: 42, CommentsEnabled: false}
		svc.On("NewPost", ctx, mock.MatchedBy(func(p post.Post) bool {
			return p.CommentsEnabled == false
		})).Return(expected, nil)

		res, err := r.CreatePost(ctx, model.CreatePostInput{Title: "T", Body: "B", CommentsEnabled: &f})
		require.NoError(t, err)
		assert.False(t, res.CommentsEnabled)
	})

	t.Run("service error", func(t *testing.T) {
		svc := new(mockService)
		r := &mutationResolver{newTestResolver(svc)}
		ctx := WithUserID(context.Background(), 42)

		svc.On("NewPost", ctx, mock.Anything).Return(post.Post{}, errs.NewValidation("title", "empty"))

		_, err := r.CreatePost(ctx, model.CreatePostInput{Title: "", Body: "B"})
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})
}

func TestSetPostCommentsEnabled(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		svc := new(mockService)
		r := &mutationResolver{newTestResolver(svc)}
		_, err := r.SetPostCommentsEnabled(context.Background(), "1", true)
		assert.True(t, errors.Is(err, errs.ErrForbidden))
	})

	t.Run("invalid id", func(t *testing.T) {
		svc := new(mockService)
		r := &mutationResolver{newTestResolver(svc)}
		ctx := WithUserID(context.Background(), 42)
		_, err := r.SetPostCommentsEnabled(ctx, "abc", true)
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("success", func(t *testing.T) {
		svc := new(mockService)
		r := &mutationResolver{newTestResolver(svc)}
		ctx := WithUserID(context.Background(), 42)

		expected := post.Post{ID: 1, CommentsEnabled: false}
		svc.On("SetCommentsEnabled", ctx, uint64(1), uint64(42), false).Return(expected, nil)

		res, err := r.SetPostCommentsEnabled(ctx, "1", false)
		require.NoError(t, err)
		assert.False(t, res.CommentsEnabled)
	})

	t.Run("forbidden from service", func(t *testing.T) {
		svc := new(mockService)
		r := &mutationResolver{newTestResolver(svc)}
		ctx := WithUserID(context.Background(), 42)
		svc.On("SetCommentsEnabled", ctx, uint64(1), uint64(42), true).Return(post.Post{}, errs.ErrForbidden)

		_, err := r.SetPostCommentsEnabled(ctx, "1", true)
		assert.True(t, errors.Is(err, errs.ErrForbidden))
	})
}

func TestCreateComment(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		svc := new(mockService)
		r := &mutationResolver{newTestResolver(svc)}
		_, err := r.CreateComment(context.Background(), model.CreateCommentInput{PostID: "1", Text: "x"})
		assert.True(t, errors.Is(err, errs.ErrForbidden))
	})

	t.Run("invalid post id", func(t *testing.T) {
		svc := new(mockService)
		r := &mutationResolver{newTestResolver(svc)}
		ctx := WithUserID(context.Background(), 42)
		_, err := r.CreateComment(ctx, model.CreateCommentInput{PostID: "abc", Text: "x"})
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("invalid parent id", func(t *testing.T) {
		svc := new(mockService)
		r := &mutationResolver{newTestResolver(svc)}
		ctx := WithUserID(context.Background(), 42)
		bad := "abc"
		_, err := r.CreateComment(ctx, model.CreateCommentInput{PostID: "1", ParentID: &bad, Text: "x"})
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("success root", func(t *testing.T) {
		svc := new(mockService)
		r := &mutationResolver{newTestResolver(svc)}
		ctx := WithUserID(context.Background(), 42)

		expected := comment.Comment{ID: 100, PostID: 1, AuthorID: 42, Text: "hello"}
		svc.On("NewComment", ctx, comment.Comment{PostID: 1, AuthorID: 42, Text: "hello"}).Return(expected, nil)

		res, err := r.CreateComment(ctx, model.CreateCommentInput{PostID: "1", Text: "hello"})
		require.NoError(t, err)
		assert.Equal(t, "100", res.ID)
		assert.Nil(t, res.ParentID)
	})

	t.Run("success reply", func(t *testing.T) {
		svc := new(mockService)
		r := &mutationResolver{newTestResolver(svc)}
		ctx := WithUserID(context.Background(), 42)
		parent := "10"

		expected := comment.Comment{ID: 100, PostID: 1, ParentID: 10, AuthorID: 42, Text: "reply"}
		svc.On("NewComment", ctx, comment.Comment{PostID: 1, ParentID: 10, AuthorID: 42, Text: "reply"}).Return(expected, nil)

		res, err := r.CreateComment(ctx, model.CreateCommentInput{PostID: "1", ParentID: &parent, Text: "reply"})
		require.NoError(t, err)
		assert.Equal(t, "10", *res.ParentID)
	})

	t.Run("service error", func(t *testing.T) {
		svc := new(mockService)
		r := &mutationResolver{newTestResolver(svc)}
		ctx := WithUserID(context.Background(), 42)

		svc.On("NewComment", ctx, mock.Anything).Return(comment.Comment{}, errs.NewValidation("text", "empty"))

		_, err := r.CreateComment(ctx, model.CreateCommentInput{PostID: "1", Text: ""})
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})
}

func TestPostComments(t *testing.T) {
	svc := new(mockService)
	r := &postResolver{newTestResolver(svc)}
	ctx := context.Background()
	obj := &model.Post{ID: "5"}

	t.Run("invalid post id", func(t *testing.T) {
		bad := &model.Post{ID: "abc"}
		_, err := r.Comments(ctx, bad, nil, nil)
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("success", func(t *testing.T) {
		svc.On("CommentTree", ctx, uint64(5), 10, uint64(0)).Return([]comment.Comment{
			{ID: 1, Text: "first"},
		}, uint64(2), nil)

		res, err := r.Comments(ctx, obj, nil, nil)
		require.NoError(t, err)
		assert.Len(t, res.Items, 1)
		assert.True(t, res.PageInfo.HasNextPage)
	})

	t.Run("with after cursor", func(t *testing.T) {
		svc2 := new(mockService)
		r2 := &postResolver{newTestResolver(svc2)}
		after := "10"
		svc2.On("CommentTree", ctx, uint64(5), 10, uint64(10)).Return([]comment.Comment{}, uint64(0), nil)

		_, err := r2.Comments(ctx, obj, nil, &after)
		require.NoError(t, err)
	})
}

func TestPosts(t *testing.T) {
	svc := new(mockService)
	r := &queryResolver{newTestResolver(svc)}
	ctx := context.Background()

	t.Run("empty", func(t *testing.T) {
		svc.On("Posts", ctx, 10, uint64(0)).Return([]post.Post{}, uint64(0), nil)

		res, err := r.Posts(ctx, nil, nil)
		require.NoError(t, err)
		assert.Empty(t, res.Items)
		assert.False(t, res.PageInfo.HasNextPage)
		assert.Nil(t, res.PageInfo.EndCursor)
	})

	t.Run("with next", func(t *testing.T) {
		svc2 := new(mockService)
		r2 := &queryResolver{newTestResolver(svc2)}
		svc2.On("Posts", ctx, 10, uint64(0)).Return([]post.Post{
			{ID: 1, Title: "T1"},
		}, uint64(1), nil)

		res, err := r2.Posts(ctx, nil, nil)
		require.NoError(t, err)
		assert.Len(t, res.Items, 1)
		assert.True(t, res.PageInfo.HasNextPage)
		assert.Equal(t, "1", *res.PageInfo.EndCursor)
	})

	t.Run("service error", func(t *testing.T) {
		svc3 := new(mockService)
		r3 := &queryResolver{newTestResolver(svc3)}
		svc3.On("Posts", ctx, 10, uint64(0)).Return(nil, uint64(0), errors.New("db error"))

		_, err := r3.Posts(ctx, nil, nil)
		assert.Error(t, err)
	})
}

func TestPost(t *testing.T) {
	svc := new(mockService)
	r := &queryResolver{newTestResolver(svc)}
	ctx := context.Background()

	t.Run("invalid id", func(t *testing.T) {
		_, err := r.Post(ctx, "abc")
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("not found returns nil", func(t *testing.T) {
		svc.On("Post", ctx, uint64(999)).Return(post.Post{}, errs.ErrNotFound)

		res, err := r.Post(ctx, "999")
		require.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("success", func(t *testing.T) {
		svc2 := new(mockService)
		r2 := &queryResolver{newTestResolver(svc2)}
		svc2.On("Post", ctx, uint64(1)).Return(post.Post{ID: 1, Title: "T"}, nil)

		res, err := r2.Post(ctx, "1")
		require.NoError(t, err)
		assert.Equal(t, "1", res.ID)
	})

	t.Run("service error", func(t *testing.T) {
		svc3 := new(mockService)
		r3 := &queryResolver{newTestResolver(svc3)}
		svc3.On("Post", ctx, uint64(1)).Return(post.Post{}, errors.New("db error"))

		_, err := r3.Post(ctx, "1")
		assert.Error(t, err)
	})
}

func TestCommentAdded(t *testing.T) {
	svc := new(mockService)
	r := &subscriptionResolver{newTestResolver(svc)}

	t.Run("invalid id", func(t *testing.T) {
		_, err := r.CommentAdded(context.Background(), "abc")
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})

	t.Run("receives published comment", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ch, err := r.CommentAdded(ctx, "1")
		require.NoError(t, err)
		require.NotNil(t, ch)

		r.broker.publish(comment.Comment{ID: 100, PostID: 1, Text: "hi"})

		select {
		case c := <-ch:
			assert.Equal(t, "100", c.ID)
			assert.Equal(t, "hi", c.Text)
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for comment")
		}
	})

	t.Run("ctx cancel closes channel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ch, err := r.CommentAdded(ctx, "1")
		require.NoError(t, err)

		cancel()
		select {
		case _, ok := <-ch:
			assert.False(t, ok, "channel should be closed after ctx cancel")
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for close")
		}
	})
}

func TestParseID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		id, err := parseID("42", "id")
		require.NoError(t, err)
		assert.Equal(t, uint64(42), id)
	})
	t.Run("invalid string", func(t *testing.T) {
		_, err := parseID("abc", "id")
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})
	t.Run("zero", func(t *testing.T) {
		_, err := parseID("0", "id")
		assert.True(t, errors.Is(err, errs.ErrValidation))
	})
}

func TestCursor(t *testing.T) {
	assert.Equal(t, uint64(0), cursor(nil))
	empty := ""
	assert.Equal(t, uint64(0), cursor(&empty))
	good := "5"
	assert.Equal(t, uint64(5), cursor(&good))
	bad := "abc"
	assert.Equal(t, uint64(0), cursor(&bad))
}

func TestPageSize(t *testing.T) {
	assert.Equal(t, 10, pageSize(nil))
	n := 5
	assert.Equal(t, 5, pageSize(&n))
}

func TestAuthenticatedUser(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		_, err := authenticatedUser(context.Background())
		assert.True(t, errors.Is(err, errs.ErrForbidden))
	})
	t.Run("zero", func(t *testing.T) {
		ctx := WithUserID(context.Background(), 0)
		_, err := authenticatedUser(ctx)
		assert.True(t, errors.Is(err, errs.ErrForbidden))
	})
	t.Run("ok", func(t *testing.T) {
		ctx := WithUserID(context.Background(), 42)
		id, err := authenticatedUser(ctx)
		require.NoError(t, err)
		assert.Equal(t, uint64(42), id)
	})
}

func TestPostModel(t *testing.T) {
	m := postModel(post.Post{ID: 1, Title: "T", Body: "B", AuthorID: 2, CommentsEnabled: true})
	assert.Equal(t, "1", m.ID)
	assert.Equal(t, "2", m.AuthorID)
	assert.True(t, m.CommentsEnabled)
}

func TestCommentModel(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		m := commentModel(comment.Comment{ID: 1, PostID: 2, AuthorID: 3, Text: "x"})
		assert.Equal(t, "1", m.ID)
		assert.Nil(t, m.ParentID)
	})
	t.Run("reply", func(t *testing.T) {
		m := commentModel(comment.Comment{ID: 1, PostID: 2, ParentID: 5, AuthorID: 3, Text: "x"})
		require.NotNil(t, m.ParentID)
		assert.Equal(t, "5", *m.ParentID)
	})
}

func TestPageInfo(t *testing.T) {
	t.Run("no next", func(t *testing.T) {
		info := pageInfo(0)
		assert.False(t, info.HasNextPage)
		assert.Nil(t, info.EndCursor)
	})
	t.Run("with next", func(t *testing.T) {
		info := pageInfo(5)
		assert.True(t, info.HasNextPage)
		require.NotNil(t, info.EndCursor)
		assert.Equal(t, "5", *info.EndCursor)
	})
}
