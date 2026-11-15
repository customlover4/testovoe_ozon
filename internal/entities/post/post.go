package post

type Post struct {
	ID              uint64
	Title           string
	Body            string
	AuthorID        uint64
	CommentsEnabled bool
}
