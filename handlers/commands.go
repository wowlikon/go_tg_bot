package handlers

import (
	"fmt"

	u "github.com/wowlikon/go_tg_bot/users"
	t "github.com/wowlikon/go_tg_bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Start(bot *tgbotapi.BotAPI, me *u.User, users *[]u.User) {
	var msg *tgbotapi.EditMessageTextConfig

	//Приветствие для пользователя
	if me.Status == u.Unregistered {
		me = u.NewUser(me.ID, u.Waiting, me.UserName, me.Directory)
		*users = append(*users, *me)
		msg = t.NewUpdMsg(me, fmt.Sprintf("Hello, %s", me.UserName))
	} else {
		msg = t.NewUpdMsg(me, "Already exist")
	}
	t.USend(bot, me, msg)
}

func UserList(bot *tgbotapi.BotAPI, me *u.User, users *[]u.User) {
	var msg *tgbotapi.EditMessageTextConfig

	//Команда только для админов
	if me.Status < u.Admin {
		msg := t.NewUpdMsg(me, "Access denied")
		t.USend(bot, me, msg)
		return
	}

	//Добавление кнопок для перехода
	ikb := tgbotapi.NewInlineKeyboardMarkup()
	kb := make([][]tgbotapi.InlineKeyboardButton, 0, len(*users))

	for _, user := range *users {
		txt := fmt.Sprintf("%s (%s)", user.UserName, user.Status)
		//ikbRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonURL(txt, fmt.Sprintf("tg://openmessage?user_id=%d", user.ID)))
		ikbRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(txt, fmt.Sprintf("user.%d", user.ID)))
		kb = append(kb, ikbRow)
	}

	msg = t.NewUpdMsg(me, "Here are the users:")
	ikb.InlineKeyboard = kb
	msg.ReplyMarkup = &ikb
	msg.ParseMode = "MarkdownV2"
	t.USend(bot, me, msg)
}

func Status(bot *tgbotapi.BotAPI, me *u.User) {
	msg := t.NewUpdMsg(me, fmt.Sprintf("You're status: %s", me.Status))
	t.USend(bot, me, msg)
}

func Help(bot *tgbotapi.BotAPI, me *u.User) {
	msg := t.NewUpdMsg(me, "To Do Later")
	t.USend(bot, me, msg)
}
