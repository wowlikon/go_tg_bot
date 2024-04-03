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

func USend(bot tgbotapi.BotAPI, emsg *tgbotapi.EditMessageTextConfig) {
	_, err := bot.Send(*emsg)
	if err != nil {
		msg := tgbotapi.NewMessage(emsg.ChatID, emsg.Text)
		msg.ReplyMarkup = emsg.ReplyMarkup
		bot.Send(msg)
	}
}

func NoCmd(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(u.FindID(update), "Error 404 command not found :(")
	bot.Send(msg)
}
