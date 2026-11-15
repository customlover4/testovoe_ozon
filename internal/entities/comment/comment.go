package comment

type Comment struct {
	ID       uint64
	PostID   uint64
	ParentID uint64
	AuthorID uint64
	Text     string
}
