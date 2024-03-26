package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func GetIndex(s []string, e string) int {
	for i := range s {
		if s[i] == e {
			return i
		}
	}
	return -1
}

func main() {
	//Проверка аргументов запуска
	args := os.Args
	fmt.Println(args)
	if (GetIndex(args, "-h") != -1) || (GetIndex(args, "--help") != -1) {
		fmt.Printf("Usage: %s [arguments]\n", args[0])
		fmt.Println("\t-h --help  | help information")
		fmt.Println("\t-d --debug | enable debug info")
		return
	}

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
		}
		messageJSON, err := json.MarshalIndent(
			update.Message, "", "  ",
		)
		if err != nil {
			msg := tgbotapi.NewMessage(
				update.Message.Chat.ID,
				fmt.Sprintf(
					"Error marshaling message to JSON: \n%s", err,
				),
			)
			bot.Send(msg)
			continue
		}
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID, fmt.Sprintf("```json\n%s\n```", messageJSON),
		)
		msg.ParseMode = "MarkdownV2"
		bot.Send(msg)

		//Проверка типов сообщений
		if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {
			if strings.HasPrefix(update.Message.Text, "/") {
				parts := strings.Split(update.Message.Text, " ")
				switch parts[0] {
				case "/start":
					start(bot, parts, update)
				case "/help":
					help(bot, update)
				default:
					nocmd(bot, update)
				}
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "not cmd")
				bot.Send(msg)
			}
		}
	}
}

func start(bot *tgbotapi.BotAPI, parts []string, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Hello, %s", parts))
	bot.Send(msg)
}

func help(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "TODO")
	bot.Send(msg)
}

func nocmd(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error 404 command not found :(")
	bot.Send(msg)
}
