package users

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type User struct {
	ID          int64  `json:"id"`
	Status      Access `json:"status"`
	UserName    string `json:"name"`
	Directory   string `json:"directory"`
	EditMessage int64  `json:"edit_msg"`
}

func NewUser(id int64, status Access, name, directory string) User {
	return User{
		ID:          id,
		Status:      status,
		UserName:    name,
		Directory:   directory,
		EditMessage: 0,
	}
}

func FindUser(users *[]User, id int64) User {
	for _, user := range *users {
		if user.ID == id {
			return user
		}
	}
	return User{id, Unregistered, "Unknown", "~", 0}
}

func GetID(update tgbotapi.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.Message.Chat.ID
	}
	return 0
}
