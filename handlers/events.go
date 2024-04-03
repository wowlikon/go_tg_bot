package handlers

import (
	"fmt"
	"strconv"

	u "github.com/wowlikon/go_tg_bot/users"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func UserInfo(bot *tgbotapi.BotAPI, update tgbotapi.Update, debug *bool, users *[]u.User, parts []string) {
	ToID := u.FindID(update)
	other_id, err := strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		if *debug {
			msg := tgbotapi.NewMessage(
				ToID,
				fmt.Sprintf("Callback parse ID error: \n %s", err),
			)
			bot.Send(msg)
		}
		return
	}

	other := u.FindUser(users, other_id)
	me := u.FindUser(users, ToID)

	if me.Status < u.Admin {
		msg := tgbotapi.NewMessage(ToID, "Permission denied")
		bot.Send(msg)
		return
	}

	//Текстовая информация
	msg := tgbotapi.NewMessage(
		ToID,
		fmt.Sprintf("Username: %s\nStatus: %s", other.UserName, other.Status),
	)

	//Добавление клавиш управления
	var ikbRow []tgbotapi.InlineKeyboardButton
	ikb := tgbotapi.NewInlineKeyboardMarkup()
	kb := make([][]tgbotapi.InlineKeyboardButton, 0, len(*users))

	//Перейти в профиль
	ikbRow = tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("Profile", fmt.Sprintf("tg://openmessage?user_id=%d", other_id)),
	)
	kb = append(kb, ikbRow)

	//Установить статус
	ikbRow = tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Set status", fmt.Sprintf("select.%d", other_id)),
	)
	kb = append(kb, ikbRow)

	//Вернуться к списку
	ikbRow = tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Back", "users"),
	)
	kb = append(kb, ikbRow)

	ikb.InlineKeyboard = kb
	msg.ReplyMarkup = ikb
	bot.Send(msg)
}

func SelectStatus(bot *tgbotapi.BotAPI, update tgbotapi.Update, debug *bool, users *[]u.User, parts []string) {
	ToID := u.FindID(update)
	other_id, err := strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		if *debug {
			msg := tgbotapi.NewMessage(
				ToID,
				fmt.Sprintf("Callback parse ID error: \n %s", err),
			)
			bot.Send(msg)
		}
		return
	}

	other := u.FindUser(users, other_id)
	me := u.FindUser(users, ToID)

	//Добавление клавиш управления
	ikb := tgbotapi.NewInlineKeyboardMarkup()
	kb := make([][]tgbotapi.InlineKeyboardButton, 0, len(*users))

	if me.Status < u.Admin {
		msg := tgbotapi.NewMessage(ToID, "Permission denied")
		bot.Send(msg)
		return
	}

	for _, v := range u.AccessList() {
		if v == 0 {
			continue
		}

		if v != u.SU {
			ikbRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(v.String(), fmt.Sprintf("set.%s.%d", parts[1], v)))
			kb = append(kb, ikbRow)
		}
	}

	if me.Status == u.SU {
		ikbRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Transfer SU", fmt.Sprintf("transferq.%s", parts[1])))
		kb = append(kb, ikbRow)
	}

	msg := tgbotapi.NewMessage(
		ToID,
		fmt.Sprintf("Select %s's access level:", other.UserName),
	)
	ikb.InlineKeyboard = kb
	msg.ReplyMarkup = ikb
	bot.Send(msg)
}

func SetStatus(bot *tgbotapi.BotAPI, update tgbotapi.Update, debug *bool, users *[]u.User, parts []string) {
	ToID := u.FindID(update)
	other_id, err := strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		if *debug {
			msg := tgbotapi.NewMessage(
				ToID,
				fmt.Sprintf("Callback parse ID error: \n %s", err),
			)
			bot.Send(msg)
		}
		return
	}

	status_id, err := strconv.Atoi(parts[2])
	if err != nil {
		if *debug {
			msg := tgbotapi.NewMessage(
				ToID,
				fmt.Sprintf("Callback parse ID error: \n %s", err),
			)
			bot.Send(msg)
		}
		return
	}

	if status_id == 0 {
		msg := tgbotapi.NewMessage(ToID, "Zero status error")
		bot.Send(msg)
		return
	}

	me := u.FindUser(users, ToID)
	if me.Status < u.Admin {
		msg := tgbotapi.NewMessage(ToID, "Permission denied")
		bot.Send(msg)
		return
	}

	name := ""
	for i, user := range *users {
		if user.ID == other_id {
			if (*users)[i].Status >= me.Status {
				msg := tgbotapi.NewMessage(ToID, "Permission denied")
				bot.Send(msg)
				return
			}
			name = user.UserName
			(*users)[i].Status = u.Access(status_id)
			break
		}
	}

	msg := tgbotapi.NewMessage(
		ToID,
		fmt.Sprintf("%s now %s", name, u.AccessList()[status_id]),
	)
	bot.Send(msg)
}

func Transferq(bot *tgbotapi.BotAPI, update tgbotapi.Update, debug *bool, users *[]u.User, parts []string) {
	ToID := u.FindID(update)
	me := u.FindUser(users, ToID)
	other_id, err := strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		if *debug {
			msg := tgbotapi.NewMessage(
				ToID,
				fmt.Sprintf("Callback parse ID error: \n %s", err),
			)
			bot.Send(msg)
		}
		return
	}

	if me.Status != u.SU {
		msg := tgbotapi.NewMessage(ToID, "Permission denied")
		bot.Send(msg)
		return
	}

	//Добавление клавиш управления
	var ikbRow []tgbotapi.InlineKeyboardButton
	ikb := tgbotapi.NewInlineKeyboardMarkup()
	kb := make([][]tgbotapi.InlineKeyboardButton, 0, len(*users))

	name := ""
	for _, user := range *users {
		if user.ID == other_id {
			name = user.UserName
			break
		}
	}

	ikbRow = tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Yes", fmt.Sprintf("transfer.%s", parts[1])))
	kb = append(kb, ikbRow)

	ikbRow = tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("No", "users"))
	kb = append(kb, ikbRow)

	msg := tgbotapi.NewMessage(
		ToID,
		fmt.Sprintf(
			"Do you want to transfer super user access to %s\n(You lost own access and become administator)",
			name,
		),
	)

	ikb.InlineKeyboard = kb
	msg.ReplyMarkup = ikb
	bot.Send(msg)
}

func Transfer(bot *tgbotapi.BotAPI, update tgbotapi.Update, debug *bool, users *[]u.User, parts []string) {
	ToID := u.FindID(update)
	me := u.FindUser(users, ToID)
	other_id, err := strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		if *debug {
			msg := tgbotapi.NewMessage(
				ToID,
				fmt.Sprintf("Callback parse ID error: \n %s", err),
			)
			bot.Send(msg)
		}
		return
	}

	if me.Status != u.SU {
		msg := tgbotapi.NewMessage(ToID, "Permission denied")
		bot.Send(msg)
		return
	}

	new_su := ""
	for _, user := range *users {
		if user.ID == other_id {
			new_su = me.UserName
			me.Status = u.Admin
			user.Status = u.SU
			break
		}
	}

	if new_su != "" {
		msg := tgbotapi.NewMessage(
			ToID,
			fmt.Sprintf("%s is SU\nNow you administrator", new_su),
		)
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(ToID, "Error")
		bot.Send(msg)
	}
}
