package handlers

import (
	"fmt"
	"strconv"
	"strings"

	wol "github.com/linde12/gowol"
	n "github.com/wowlikon/go_lan_scanner/lib"
	u "github.com/wowlikon/go_tg_bot/users"
	t "github.com/wowlikon/go_tg_bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func UserList(bot *tgbotapi.BotAPI, us u.SelectedUser) {
	var msg *tgbotapi.EditMessageTextConfig
	me := u.GetUser(us)

	//Команда только для админов
	if me.Status < u.Admin {
		msg := t.NewUpdMsg(us, "Access denied")
		t.USend(bot, us, msg)
		return
	}

	//Добавление кнопок для перехода
	ikb := tgbotapi.NewInlineKeyboardMarkup()
	kb := make([][]tgbotapi.InlineKeyboardButton, 0, len(*us.Users))

	for _, user := range *(us.Users) {
		txt := fmt.Sprintf("%s (%s)", user.UserName, user.Status)
		//ikbRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonURL(txt, fmt.Sprintf("tg://openmessage?user_id=%d", user.ID)))
		ikbRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(txt, fmt.Sprintf("user.%d", user.ID)))
		kb = append(kb, ikbRow)
	}

	msg = t.NewUpdMsg(us, "Here are the users:")
	ikb.InlineKeyboard = kb
	msg.ReplyMarkup = &ikb
	msg.ParseMode = "MarkdownV2"
	t.USend(bot, us, msg)
}

func UserInfo(bot *tgbotapi.BotAPI, us u.SelectedUser, parts *[]string) {
	me := u.GetUser(us)
	other_id, err := strconv.ParseInt((*parts)[1], 10, 0)
	if err != nil {
		return
	}

	other := u.FindUser(us.Users, other_id, "unknown")

	if me.Status < u.Admin {
		NoPermission(bot, us)
		return
	}

	//Текстовая информация
	msg := t.NewUpdMsg(
		us, fmt.Sprintf("Username: %s\nStatus: %s", u.GetUser(other).UserName, u.GetUser(other).Status),
	)

	//Добавление клавиш управления
	var ikbRow []tgbotapi.InlineKeyboardButton
	ikb := tgbotapi.NewInlineKeyboardMarkup()
	kb := make([][]tgbotapi.InlineKeyboardButton, 0, 3)

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
	t.USend(bot, us, msg)
}

func SelectStatus(bot *tgbotapi.BotAPI, us u.SelectedUser, parts *[]string) {
	me := u.GetUser(us)
	other_id, err := strconv.ParseInt((*parts)[1], 10, 0)
	if err != nil {
		return
	}

	other := u.FindUser(us.Users, other_id, "unknown")
	if me.ID == other.ID {
		msg := t.NewUpdMsg(us, "You can't set self status")
		t.USend(bot, us, msg)
		return
	}

	//Добавление клавиш управления
	ikb := tgbotapi.NewInlineKeyboardMarkup()
	kb := make([][]tgbotapi.InlineKeyboardButton, 0, len(u.AccessList()))

	if me.Status < u.Admin {
		NoPermission(bot, us)
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
		us, fmt.Sprintf("Select %s's access level:", other.UserName),
	)
	ikb.InlineKeyboard = kb
	msg.ReplyMarkup = &ikb
	t.USend(bot, us, msg)
}

func SetStatus(bot *tgbotapi.BotAPI, us u.SelectedUser, parts *[]string) {
	me := u.GetUser(us)
	other_id, err := strconv.ParseInt((*parts)[1], 10, 0)
	if err != nil {
		return
	}

	status_id, err := strconv.Atoi((*parts)[2])
	if err != nil {
		return
	}

	if status_id == 0 {
		msg := t.NewUpdMsg(us, "Zero status error")
		t.USend(bot, us, msg)
		return
	}

	if other_id == me.ID {
		msg := t.NewUpdMsg(us, "You can't set self status")
		t.USend(bot, us, msg)
		return
	}

	if me.Status < u.Admin {
		NoPermission(bot, us)
		return
	}

	name := ""
	for i, user := range *us.Users {
		if user.ID == other_id {
			if (*us.Users)[i].Status >= me.Status {
				NoPermission(bot, us)
				return
			}
			name = user.UserName
			(*us.Users)[i].Status = u.Access(status_id)
			break
		}
	}

	msg := t.NewUpdMsg(
		us, fmt.Sprintf("%s now %s", name, u.AccessList()[status_id]),
	)
	t.USend(bot, us, msg)
}

func Transferq(bot *tgbotapi.BotAPI, us u.SelectedUser, parts *[]string) {
	me := u.GetUser(us)
	other_id, err := strconv.ParseInt((*parts)[1], 10, 0)
	if err != nil {
		return
	}

	if me.Status != u.SU {
		NoPermission(bot, us)
		return
	}

	//Добавление клавиш управления
	var ikbRow []tgbotapi.InlineKeyboardButton
	ikb := tgbotapi.NewInlineKeyboardMarkup()
	kb := make([][]tgbotapi.InlineKeyboardButton, 0, 2)

	name := ""
	for _, user := range *us.Users {
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
		us, fmt.Sprintf(
			"Do you want to transfer super user access to %s\n(You lost own access and become administator)",
			name,
		),
	)

	ikb.InlineKeyboard = kb
	msg.ReplyMarkup = &ikb
	t.USend(bot, us, msg)
}

func Transfer(bot *tgbotapi.BotAPI, us u.SelectedUser, parts *[]string) {
	me := u.GetUser(us)
	other_id, err := strconv.ParseInt((*parts)[1], 10, 0)
	if err != nil {
		return
	}

	if me.Status != u.SU {
		NoPermission(bot, us)
		return
	}

	new_su := ""
	for i, user := range *us.Users {
		if user.ID == other_id {
			new_su = user.UserName
			(*us.Users)[i].Status = u.SU
		}
		if user.ID == me.ID {
			(*us.Users)[i].Status = u.Admin
		}
	}

	if new_su != "" {
		msg := t.NewUpdMsg(
			us, fmt.Sprintf("%s is SU\nNow you administrator", new_su),
		)
		t.USend(bot, us, msg)
	} else {
		msg := t.NewUpdMsg(us, "Error")
		t.USend(bot, us, msg)
	}
}

func SetDebug(bot *tgbotapi.BotAPI, debug *bool, us u.SelectedUser, parts *[]string) {
	var ikbRow []tgbotapi.InlineKeyboardButton
	me := u.GetUser(us)

	if me.Status != u.SU {
		NoPermission(bot, us)
		return
	}

	if len(*parts) == 1 {
		*parts = append(*parts, "")
	}

	if strings.ToLower((*parts)[1]) == "on" {
		msg := t.NewUpdMsg(us, "Debug mode on!")
		t.USend(bot, us, msg)
		*debug = true
		return
	}

	if strings.ToLower((*parts)[1]) == "off" {
		msg := t.NewUpdMsg(us, "Debug mode off!")
		t.USend(bot, us, msg)
		*debug = false
		return
	}

	msg := t.NewUpdMsg(us, fmt.Sprintf("Set debug status\nNow value: %t", *debug))
	ikb := tgbotapi.NewInlineKeyboardMarkup()
	kb := make([][]tgbotapi.InlineKeyboardButton, 0, 2)

	//Установить значение
	ikbRow = tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("ON", "debug.on"),
	)
	kb = append(kb, ikbRow)

	ikbRow = tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("OFF", "debug.off"),
	)
	kb = append(kb, ikbRow)

	ikb.InlineKeyboard = kb
	msg.ReplyMarkup = &ikb
	t.USend(bot, us, msg)
}

func RequestPower(bot *tgbotapi.BotAPI, us u.SelectedUser) {
	var msg *tgbotapi.EditMessageTextConfig
	owner := u.FindSU(us.Users)
	me := u.GetUser(us)

	//Сообщение об отправке запроса
	msg = t.NewUpdMsg(us, "Permission requested.")
	t.USend(bot, us, msg)

	//Сообщение для владельца
	msg = t.NewUpdMsg(owner, fmt.Sprintf("User %s trying to power", me.UserName))
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Power", "power"),
		),
	)
	msg.ReplyMarkup = &kb
	t.USend(bot, owner, msg)
}

func Power(bot *tgbotapi.BotAPI, us u.SelectedUser, conf t.Configuration) {
	if packet, err := wol.NewMagicPacket(conf.Device.MAC); err == nil {
		packet.Send("255.255.255.255")          // send to broadcast
		packet.SendPort("255.255.255.255", "7") // specify receiving port
	}
}

func Config(bot *tgbotapi.BotAPI, us u.SelectedUser) {
	var ikbRow []tgbotapi.InlineKeyboardButton
	me := u.GetUser(us)

	if me.Status != u.SU {
		NoPermission(bot, us)
		return
	}

	msg := t.NewUpdMsg(us, "Here you can configure this bot")
	ikb := tgbotapi.NewInlineKeyboardMarkup()
	kb := make([][]tgbotapi.InlineKeyboardButton, 0, 2)

	//Установить значение
	ikbRow = tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Set Debug", "debug"),
	)
	kb = append(kb, ikbRow)

	ikbRow = tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Scan LAN", "scan"),
	)
	kb = append(kb, ikbRow)

	ikb.InlineKeyboard = kb
	msg.ReplyMarkup = &ikb
	t.USend(bot, us, msg)
}

func ScanList(bot *tgbotapi.BotAPI, us u.SelectedUser, conf t.Configuration) {
	var ikbRow []tgbotapi.InlineKeyboardButton
	me := u.GetUser(us)

	if me.Status != u.SU {
		NoPermission(bot, us)
		return
	}

	msg := t.NewUpdMsg(us, "Select device for WOL and control")
	ikb := tgbotapi.NewInlineKeyboardMarkup()
	kb := make([][]tgbotapi.InlineKeyboardButton, 0, len(conf.Devices)+1)

	//Выбрать из найденого
	for _, device := range conf.Devices {
		ikbRow = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				device.Name,
				fmt.Sprintf("device.%s", device.MAC),
			),
		)
		kb = append(kb, ikbRow)
	}

	//Запустить новое сканирование
	ikbRow = tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Run new scan ↻", "rescan"),
	)
	kb = append(kb, ikbRow)

	ikb.InlineKeyboard = kb
	msg.ReplyMarkup = &ikb
	t.USend(bot, us, msg)
}

func Rescan(bot *tgbotapi.BotAPI, us u.SelectedUser, conf t.Configuration) {
	conf.Devices, _ = t.FilterDevices(n.Scan("192.168.0.0/24", n.PortList))
	ScanList(bot, us, conf)
}

func SetDevice(bot *tgbotapi.BotAPI, us u.SelectedUser, conf t.Configuration, parts *[]string) {

	if len(*parts) == 1 {
		*parts = append(*parts, "")
	}

	for _, device := range conf.Devices {
		if device.MAC == (*parts)[1] {
			conf.Device = device
			msg := t.NewUpdMsg(us, "Device value is set.")
			t.USend(bot, us, msg)
			return
		}
	}
	msg := t.NewUpdMsg(us, "Error 404 device not found :(")
	t.USend(bot, us, msg)
}
