package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	h "github.com/wowlikon/go_tg_bot/handlers"
	u "github.com/wowlikon/go_tg_bot/users"
	t "github.com/wowlikon/go_tg_bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var debug, key_used bool

func main() {
	var users []u.User

	//Проверка аргументов запуска
	args := os.Args
	if (t.GetIndex(args, "-h") != -1) || (t.GetIndex(args, "--help") != -1) {
		fmt.Printf("Usage: %s [arguments]\n", args[0])
		fmt.Println("\t-h --help  | help information")
		fmt.Println("\t-d --debug | enable debug info")
		return
	}

	debug = (t.GetIndex(args, "-d") != -1) || (t.GetIndex(args, "--debug") != -1)

	//Загружаем .env
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}

	key_len, _ := strconv.Atoi(os.Getenv("KEY_LENGTH"))
	key, _ := t.GenerateKey(key_len)
	fmt.Printf("Admin key: %s\n", key)

	//Создаем бота
	fmt.Println("Starting bot")
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Bot @%s is online in ", bot.Self.UserName)
	if debug {
		fmt.Println("debug mode")
	} else {
		fmt.Println("standart mode")
	}

	//Устанавливаем время обновления
	upd := tgbotapi.NewUpdate(0)
	upd.Timeout = 60

	//Получаем обновления от бота
	for update := range bot.GetUpdatesChan(upd) {

		ToID := u.GetID(update)
		if ToID == 0 {
			continue
		}

		//Вывод данных о сообщении
		if debug {
			updateJSON, err := json.MarshalIndent(
				update, "", "  ",
			)
			if err != nil {
				msg := tgbotapi.NewMessage(
					ToID, fmt.Sprintf(
						"Error marshaling update to JSON: \n%s", err,
					),
				)
				bot.Send(msg)
				continue
			}
			msg := tgbotapi.NewMessage(
				ToID, fmt.Sprintf("```json\n%s\n```", updateJSON),
			)
			msg.ParseMode = "MarkdownV2"
			bot.Send(msg)
		}

		//Проверка типа на сообщение
		if update.Message != nil {
			var msg tgbotapi.MessageConfig

			//Проверка на одноразовый ключ доступа
			if (update.Message.Text == key) && !key_used {
				userName := update.Message.From.UserName
				id := update.Message.From.ID
				key_used = true
				idx := -1
				for userid, user := range users {
					if user.ID == id {
						idx = userid
						break
					}
				}

				if idx == -1 {
					users = append(users, u.NewUser(id, u.SU, userName, "~"))
				} else {
					users[idx].Status = u.SU
				}

				//Приветствие для суперадминистратора
				msg = tgbotapi.NewMessage(ToID, fmt.Sprintf("```welcome $sudo hello_world --admin %s```", userName))
				msg.ParseMode = "MarkdownV2"
				bot.Send(msg)
				continue
			}

			//Проверка команды
			if strings.HasPrefix(update.Message.Text, "/") {
				parts := strings.Split(update.Message.Text, " ")
				switch parts[0] {
				case "/start":
					h.Start(bot, update, &users)
				case "/users":
					h.UserList(bot, update, &debug, &users)
				case "/status":
					h.Status(bot, update, &users)
				case "/debug":
					h.ToggleDebug(bot, update, &debug, &users, parts)
				case "/help":
					h.Help(bot, update)
				default:
					t.NoCmd(bot, update)
				}
			} else {
				//Если просто текст
				msg := tgbotapi.NewMessage(ToID, "not command TODO")
				bot.Send(msg)
			}

			//Проверка типа на событие кнопки
		} else if update.CallbackQuery != nil {
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			parts := strings.Split(update.CallbackQuery.Data, ".")
			if _, err := bot.Request(callback); err != nil {
				msg := tgbotapi.NewMessage(
					ToID, fmt.Sprintf("Callback error: \n%s", err),
				)
				bot.Send(msg)
				continue
			}

			//Проверка события
			switch parts[0] {
			case "user":
				h.UserInfo(bot, update, &debug, &users, parts)
			case "users":
				h.UserList(bot, update, &debug, &users)
			case "select":
				h.SelectStatus(bot, update, &debug, &users, parts)
			case "set":
				h.SetStatus(bot, update, &debug, &users, parts)
			case "transferq":
				h.Transferq(bot, update, &debug, &users, parts)
			case "transfer":
				h.Transfer(bot, update, &debug, &users, parts)
			default:
				t.NoCmd(bot, update)
			}
		}
	}
}
