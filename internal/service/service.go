package service

type storage interface {
	auth
	postService
}

type service struct {
	stor   storage
	secret []byte
}

func New(s storage, secret string) *service {
	return &service{
		stor:   s,
		secret: []byte(secret),
	}
}
