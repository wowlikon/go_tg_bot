package handlers

import (
	"encoding/json"
	"fmt"
	"strings"

	u "github.com/wowlikon/go_tg_bot/users"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Start(bot *tgbotapi.BotAPI, update tgbotapi.Update, users *[]u.User) {
	var msg tgbotapi.MessageConfig
	ToID := u.GetID(update)

	//Проверка на повторный запуск
	userName := update.Message.From.UserName
	id := update.Message.From.ID
	exist := (u.FindUser(users, id).Status != u.Unregistered)

	//Приветствие для пользователя
	if !exist {
		*users = append(*users, u.User{ID: id, Status: u.Waiting, UserName: userName, Directory: "~", EditMessage: 0})
		msg = tgbotapi.NewMessage(ToID, fmt.Sprintf("Hello, %s", userName))
	} else {
		msg = tgbotapi.NewMessage(ToID, "Already exist")
	}
	bot.Send(msg)
}

func UserList(bot *tgbotapi.BotAPI, update tgbotapi.Update, debug *bool, users *[]u.User) {
	var msg tgbotapi.MessageConfig

	//Команда только для админов
	ToID := u.GetID(update)
	if ToID == 0 {
		return
	}
	if u.FindUser(users, ToID).Status < u.Admin {
		msg := tgbotapi.NewMessage(ToID, "Access denied")
		bot.Send(msg)
		return
	}

	//Вывод списка пользователей
	if *debug {
		usersJSON, err := json.MarshalIndent(
			*users, "", "  ",
		)
		if err != nil {
			msg := tgbotapi.NewMessage(
				ToID,
				fmt.Sprintf(
					"Error marshaling users to JSON: \n%s", err,
				),
			)
			bot.Send(msg)
		}
		msg = tgbotapi.NewMessage(
			ToID, fmt.Sprintf("```json\n%s\n```", usersJSON),
		)
	} else {
		//Добавление кнопок для перехода
		ikb := tgbotapi.NewInlineKeyboardMarkup()
		kb := make([][]tgbotapi.InlineKeyboardButton, 0, len(*users))

		for _, user := range *users {
			txt := fmt.Sprintf("%s (%s)", user.UserName, user.Status)
			//ikbRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonURL(txt, fmt.Sprintf("tg://openmessage?user_id=%d", user.ID)))
			ikbRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(txt, fmt.Sprintf("user.%d", user.ID)))
			kb = append(kb, ikbRow)
		}

		msg = tgbotapi.NewMessage(
			ToID,
			"Here are the users:",
		)
		ikb.InlineKeyboard = kb
		msg.ReplyMarkup = ikb
	}
	msg.ParseMode = "MarkdownV2"
	bot.Send(msg)
}

func ToggleDebug(bot *tgbotapi.BotAPI, update tgbotapi.Update, debug *bool, users *[]u.User, parts []string) {
	ToID := u.GetID(update)
	my_status := u.FindUser(users, ToID).Status
	if my_status != u.SU {
		msg := tgbotapi.NewMessage(ToID, "Access denied!")
		bot.Send(msg)
		return
	}

	if len(parts) == 1 {
		parts = append(parts, "")
	}

	if strings.ToLower(parts[1]) == "on" {
		msg := tgbotapi.NewMessage(ToID, "Debug mode on!")
		bot.Send(msg)
		*debug = true
		return
	}

	if strings.ToLower(parts[1]) == "off" {
		msg := tgbotapi.NewMessage(ToID, "Debug mode off!")
		bot.Send(msg)
		*debug = false
		return
	}

	msg := tgbotapi.NewMessage(
		ToID,
		fmt.Sprintf("Debug: %t \n/debug [on/off]", *debug),
	)
	bot.Send(msg)
}

func Status(bot *tgbotapi.BotAPI, update tgbotapi.Update, users *[]u.User) {
	ToID := u.GetID(update)
	my_status := u.FindUser(users, ToID).Status
	msg := tgbotapi.NewMessage(
		ToID,
		fmt.Sprintf("You're status: %s", my_status),
	)
	bot.Send(msg)
}

func Help(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(u.GetID(update), "TODO")
	bot.Send(msg)
}
