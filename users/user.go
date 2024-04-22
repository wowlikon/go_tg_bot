package users

//Пользователь
type User struct {
	ID          int64  `json:"id"`
	Status      Access `json:"status"`
	UserName    string `json:"name"`
	Directory   string `json:"directory"`
	EditMessage int    `json:"edit_msg"`
}

//Ссылка на пользователя
type SelectedUser struct {
	ID       int64   `json:"id"`
	Index    int     `json:"index"`
	UserName string  `json:"name"`
	Users    *[]User `json:"users"`
}

//Создание нового пользователя
func NewUser(id int64, status Access, name, directory string) *User {
	return &User{
		ID:          id,
		Status:      status,
		UserName:    name,
		Directory:   directory,
		EditMessage: 0,
	}
}

//Пользователь не найден
func NoUser(id int64, userName string) *User {
	return &User{
		ID:          id,
		Status:      Unregistered,
		UserName:    userName,
		Directory:   "~",
		EditMessage: 0,
	}
}

//Найти пользователя по id
func FindUser(users *[]User, id int64, name string) SelectedUser {
	for idx, user := range *users {
		if user.ID == id {
			return SelectedUser{id, idx, name, users}
		}
	}
	return SelectedUser{id, -1, name, users}
}

//Получить пользователя из ссылки
func GetUser(elem SelectedUser) *User {
	if elem.Index == -1 {
		return NoUser(elem.ID, elem.UserName)
	}
	return &((*elem.Users)[elem.Index])
}
