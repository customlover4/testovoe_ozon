package user

type User struct {
	ID       uint64
	Login    string
	Password string
}

func (u User) Empty() bool {
	if u.ID == 0 && u.Login == "" && u.Password == "" {
		return true
	}
	
	return false
}