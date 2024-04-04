package users

type User struct {
	ID          int64  `json:"id"`
	Status      Access `json:"status"`
	UserName    string `json:"name"`
	Directory   string `json:"directory"`
	EditMessage int    `json:"edit_msg"`
}

func NewUser(id int64, status Access, name, directory string) *User {
	return &User{
		ID:          id,
		Status:      status,
		UserName:    name,
		Directory:   directory,
		EditMessage: 0,
	}
}

func NoUser(id int64, userName string) *User {
	return &User{
		ID:          id,
		Status:      Unregistered,
		UserName:    userName,
		Directory:   "~",
		EditMessage: 0,
	}
}

func FindUser(users *[]User, id int64, name string) *User {
	for _, user := range *users {
		if user.ID == id {
			return &user
		}
	}
	return NoUser(id, name)
}
