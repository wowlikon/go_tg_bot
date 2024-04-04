package utils

import (
	"crypto/rand"
	"encoding/hex"

	u "github.com/wowlikon/go_tg_bot/users"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

func NewUpdMsg(me *u.User, text string) *tgbotapi.EditMessageTextConfig {
	umsg := tgbotapi.NewEditMessageText(me.ID, me.EditMessage, text)
	return &umsg
}

func USend(bot *tgbotapi.BotAPI, me *u.User, emsg *tgbotapi.EditMessageTextConfig) {
	var sended tgbotapi.Message
	var err error

	sended, err = bot.Send(*emsg)
	if err != nil {
		msg := tgbotapi.NewMessage(emsg.ChatID, emsg.Text)
		msg.ReplyMarkup = emsg.ReplyMarkup
		sended, _ = bot.Send(msg)
	}

	(*me).EditMessage = sended.MessageID
}

func NoCmd(bot *tgbotapi.BotAPI, me *u.User) {
	msg := NewUpdMsg(me, "Error 404 command not found :(")
	USend(bot, me, msg)
}
