package handlers

import (
	"fmt"
	"strconv"

	u "github.com/wowlikon/go_tg_bot/users"
	t "github.com/wowlikon/go_tg_bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func UserInfo(bot *tgbotapi.BotAPI, me *u.User, users *[]u.User, parts *[]string) {
	other_id, err := strconv.ParseInt((*parts)[1], 10, 0)
	if err != nil {
		return
	}

	other := u.FindUser(users, other_id, "unknown")

	if me.Status < u.Admin {
		msg := t.NewUpdMsg(me, "Permission denied")
		t.USend(bot, me, msg)
		return
	}

	//Текстовая информация
	msg := t.NewUpdMsg(
		me, fmt.Sprintf("Username: %s\nStatus: %s", other.UserName, other.Status),
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
	msg.ReplyMarkup = &ikb
	t.USend(bot, me, msg)
}

func SelectStatus(bot *tgbotapi.BotAPI, me *u.User, users *[]u.User, parts *[]string) {
	other_id, err := strconv.ParseInt((*parts)[1], 10, 0)
	if err != nil {
		return
	}

	other := u.FindUser(users, other_id, "unknown")

	//Добавление клавиш управления
	ikb := tgbotapi.NewInlineKeyboardMarkup()
	kb := make([][]tgbotapi.InlineKeyboardButton, 0, len(*users))

	if me.Status < u.Admin {
		msg := t.NewUpdMsg(me, "Permission denied")
		t.USend(bot, me, msg)
		return
	}

	for _, v := range u.AccessList() {
		if v == 0 {
			continue
		}

		if v != u.SU {
			ikbRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(v.String(), fmt.Sprintf("set.%s.%d", (*parts)[1], v)))
			kb = append(kb, ikbRow)
		}
	}

	if me.Status == u.SU {
		ikbRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Transfer SU", fmt.Sprintf("transferq.%s", (*parts)[1])))
		kb = append(kb, ikbRow)
	}

	msg := t.NewUpdMsg(
		me, fmt.Sprintf("Select %s's access level:", other.UserName),
	)
	ikb.InlineKeyboard = kb
	msg.ReplyMarkup = &ikb
	t.USend(bot, me, msg)
}

func SetStatus(bot *tgbotapi.BotAPI, me *u.User, users *[]u.User, parts *[]string) {
	other_id, err := strconv.ParseInt((*parts)[1], 10, 0)
	if err != nil {
		return
	}

	status_id, err := strconv.Atoi((*parts)[2])
	if err != nil {
		return
	}

	if status_id == 0 {
		msg := t.NewUpdMsg(me, "Zero status error")
		t.USend(bot, me, msg)
		return
	}

	if other_id == me.ID {
		msg := t.NewUpdMsg(me, "You can't set self status")
		t.USend(bot, me, msg)
		return
	}

	if me.Status < u.Admin {
		msg := t.NewUpdMsg(me, "Permission denied")
		t.USend(bot, me, msg)
		return
	}

	name := ""
	for i, user := range *users {
		if user.ID == other_id {
			if (*users)[i].Status >= me.Status {
				msg := t.NewUpdMsg(me, "Permission denied")
				t.USend(bot, me, msg)
				return
			}
			name = user.UserName
			(*users)[i].Status = u.Access(status_id)
			break
		}
	}

	msg := t.NewUpdMsg(
		me, fmt.Sprintf("%s now %s", name, u.AccessList()[status_id]),
	)
	t.USend(bot, me, msg)
}

func Transferq(bot *tgbotapi.BotAPI, me *u.User, users *[]u.User, parts *[]string) {
	other_id, err := strconv.ParseInt((*parts)[1], 10, 0)
	if err != nil {
		return
	}

	if me.Status != u.SU {
		msg := t.NewUpdMsg(me, "Permission denied")
		t.USend(bot, me, msg)
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

	ikbRow = tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Yes", fmt.Sprintf("transfer.%s", (*parts)[1])))
	kb = append(kb, ikbRow)

	ikbRow = tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("No", "users"))
	kb = append(kb, ikbRow)

	msg := t.NewUpdMsg(
		me, fmt.Sprintf(
			"Do you want to transfer super user access to %s\n(You lost own access and become administator)",
			name,
		),
	)

	ikb.InlineKeyboard = kb
	msg.ReplyMarkup = &ikb
	t.USend(bot, me, msg)
}

func Transfer(bot *tgbotapi.BotAPI, me *u.User, users *[]u.User, parts *[]string) {
	other_id, err := strconv.ParseInt((*parts)[1], 10, 0)
	if err != nil {
		return
	}

	if me.Status != u.SU {
		msg := t.NewUpdMsg(me, "Permission denied")
		t.USend(bot, me, msg)
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
		msg := t.NewUpdMsg(
			me, fmt.Sprintf("%s is SU\nNow you administrator", new_su),
		)
		t.USend(bot, me, msg)
	} else {
		msg := t.NewUpdMsg(me, "Error")
		t.USend(bot, me, msg)
	}
}
