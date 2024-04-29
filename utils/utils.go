package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	n "github.com/wowlikon/go_lan_scanner/lib"
	u "github.com/wowlikon/go_tg_bot/users"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const ignore = "Bad Request: message is not modified: specified new message content and reply markup are exactly the same as a current content and reply markup of the message"

func GenerateKey(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GetIndex(s []string, e string) int {
	for i := range s {
		if s[i] == e {
			return i
		}
	}
	return -1
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

func GetFrom(update tgbotapi.Update) *tgbotapi.User {
	if update.Message != nil {
		return update.Message.From
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.Message.From
	}
	return nil
}

func FilterDevices(devices []n.Device, err error) ([]n.Device, error) {
	if err != nil {
		return nil, err
	}

	var filtered []n.Device
	for _, device := range devices {
		if device.MAC != "" {
			filtered = append(filtered, device)
		}
	}
	return filtered, nil

}

func NewUpdMsg(us u.SelectedUser, text string) *tgbotapi.EditMessageTextConfig {
	me := u.GetUser(us)
	umsg := tgbotapi.NewEditMessageText(
		me.ID, me.EditMessage, text,
	)
	return &umsg
}

func USend(bot *tgbotapi.BotAPI, us u.SelectedUser, emsg *tgbotapi.EditMessageTextConfig) {
	var msgID int
	var err error

	msgID = u.GetUser(us).EditMessage
	_, err = bot.Send(*emsg)

	if err != nil {
		if err.Error() != ignore {
			fmt.Println(err, msgID)
			msg := tgbotapi.NewMessage(emsg.ChatID, emsg.Text)
			msg.ReplyMarkup = emsg.ReplyMarkup
			sended, _ := bot.Send(msg)
			msgID = sended.MessageID
		}
	}

	if us.Index == -1 {
		return
	}

	(*us.Users)[us.Index].EditMessage = msgID
}
