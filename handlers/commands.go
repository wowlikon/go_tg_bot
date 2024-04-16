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

func Status(bot *tgbotapi.BotAPI, me *u.User) {
	msg := t.NewUpdMsg(me, fmt.Sprintf("You're status: %s", me.Status))
	t.USend(bot, me, msg)
}

func Main(bot *tgbotapi.BotAPI, me *u.User) {
	var ikbRow []tgbotapi.InlineKeyboardButton
	var msg *tgbotapi.EditMessageTextConfig

	if me.Status <= u.Waiting {
		msg = t.NewUpdMsg(me, "Permision demied")
		t.USend(bot, me, msg)
		return
	}

	//Добавление кнопок для перехода
	ikb := tgbotapi.NewInlineKeyboardMarkup()
	kb := make([][]tgbotapi.InlineKeyboardButton, 0, 4)

	if me.Status >= u.Admin {
		ikbRow = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Users", "users"),
		)
	}
	if me.Status == u.SU {
		ikbRow = append(ikbRow,
			tgbotapi.NewInlineKeyboardButtonData("Config", "config"),
		)
		kb = append(kb, ikbRow)

		ikbRow = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Termial", "terminal"),
		)
	}
	kb = append(kb, ikbRow)

	ikbRow = tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Files", "files"),
	)
	kb = append(kb, ikbRow)

	ikbRow = tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Help", "Help"),
	)
	kb = append(kb, ikbRow)

	if me.Status >= u.Admin {
		ikbRow = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Wake/Shutdown", "power"),
		)
		kb = append(kb, ikbRow)
	} else {
		ikbRow = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Request Wake/Shutdown", "powerq"),
		)
		kb = append(kb, ikbRow)
	}

	msg = t.NewUpdMsg(me, fmt.Sprintf("You're status: %s", me.Status))
	ikb.InlineKeyboard = kb
	msg.ReplyMarkup = &ikb
	msg.ParseMode = "MarkdownV2"
	t.USend(bot, me, msg)
}

func Help(bot *tgbotapi.BotAPI, me *u.User) {
	var hint string
	hint += "/start - begin using bot\n"
	hint += "/main - command to interact with bot\n"
	hint += "/help - get this information\n"
	msg := t.NewUpdMsg(me, hint)
	t.USend(bot, me, msg)
}

func TODO(bot *tgbotapi.BotAPI, me *u.User) {
	msg := t.NewUpdMsg(me, "TODO")
	t.USend(bot, me, msg)
}
