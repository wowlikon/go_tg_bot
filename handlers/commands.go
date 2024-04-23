package handlers

import (
	"fmt"

	u "github.com/wowlikon/go_tg_bot/users"
	t "github.com/wowlikon/go_tg_bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Команда начала взаимодействия
func Start(bot *tgbotapi.BotAPI, us u.SelectedUser) {
	var msg *tgbotapi.EditMessageTextConfig
	me := u.GetUser(us)

	//Приветствие для пользователя
	if me.Status == u.Unregistered {
		me = u.NewUser(me.ID, u.Waiting, me.UserName, me.Directory)
		*us.Users = append(*us.Users, *me)
		msg = t.NewUpdMsg(us, fmt.Sprintf("Hello, %s", me.UserName))
	} else {
		msg = t.NewUpdMsg(us, "Already exist")
	}
	t.USend(bot, us, msg)
}

// Вывести статус пользователя
func Status(bot *tgbotapi.BotAPI, us u.SelectedUser) {
	me := u.GetUser(us)

	msg := t.NewUpdMsg(us, fmt.Sprintf("You're status: %s", me.Status))
	t.USend(bot, us, msg)
}

// Основная команда бота
func Main(bot *tgbotapi.BotAPI, us u.SelectedUser) {
	var ikbRow []tgbotapi.InlineKeyboardButton
	var msg *tgbotapi.EditMessageTextConfig
	me := u.GetUser(us)

	if me.Status <= u.Waiting {
		NoPermission(bot, us)
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
		tgbotapi.NewInlineKeyboardButtonData("Help", "help"),
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

	msg = t.NewUpdMsg(us, fmt.Sprintf("You're status: %s", me.Status))
	ikb.InlineKeyboard = kb
	msg.ReplyMarkup = &ikb
	msg.ParseMode = "MarkdownV2"
	t.USend(bot, us, msg)
}

// Команда для вывода подсказок
func Help(bot *tgbotapi.BotAPI, us u.SelectedUser) {
	var hint string
	hint += "/start - begin using bot\n"
	hint += "/main - command to interact with bot\n"
	hint += "/help - get this information\n"
	msg := t.NewUpdMsg(us, hint)
	t.USend(bot, us, msg)
}

// Если функция не доделана
func TODO(bot *tgbotapi.BotAPI, us u.SelectedUser) {
	msg := t.NewUpdMsg(us, "TODO")
	t.USend(bot, us, msg)
}

// Если команда не найдена
func NoCmd(bot *tgbotapi.BotAPI, us u.SelectedUser) {
	msg := t.NewUpdMsg(us, "Error 404 command not found :(")
	t.USend(bot, us, msg)
}

// Если недостаточно прав
func NoPermission(bot *tgbotapi.BotAPI, us u.SelectedUser) {
	msg := t.NewUpdMsg(us, "Permission denied :(")
	t.USend(bot, us, msg)
}
