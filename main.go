package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/joho/godotenv"
)

func main() {
	//Загружаем .env
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}

	//Создаем бота
	fmt.Println("Starting bot")
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Bot @%s is online\n", bot.Self.UserName)

	//Устанавливаем время обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	//Получаем обновления от бота
	for update := range bot.GetUpdatesChan(u) {

		//Вывод данных о сообщении
		if update.Message == nil {
			continue
		} else {
			messageJSON, err := json.MarshalIndent(
				update.Message, "", "  ",
			)
			if err != nil {
				fmt.Println("Error marshaling message to JSON:", err)
				continue
			}
			msg := tgbotapi.NewMessage(
				update.Message.Chat.ID, string(messageJSON),
			)
			bot.Send(msg)
			fmt.Println(string(messageJSON))
		}

		//Проверка типов сообщений
		if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {
		} else {
			msg := tgbotapi.NewMessage(
				update.Message.Chat.ID, "Incorrect type",
			)
			bot.Send(msg)
			continue
		}
	}
}
