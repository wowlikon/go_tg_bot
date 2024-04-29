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

// var debug, key_used bool
var conf t.Configuration

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

	conf.SetDebug((t.GetIndex(args, "-d") != -1) || (t.GetIndex(args, "--debug") != -1))

	//Загружаем .env
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}

	key_len, _ := strconv.Atoi(os.Getenv("KEY_LENGTH"))
	key, _ := t.GenerateKey(key_len)
	fmt.Printf("Admin key: %s\n", key)
	conf.SetKey(key)

	//Создаем бота
	fmt.Println("Starting bot")
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}

	fmt.Printf("Bot @%s is online in ", bot.Self.UserName)
	if conf.Debug {
		fmt.Println("debug mode")
	} else {
		fmt.Println("standart mode")
	}

	//Устанавливаем время обновления
	upd := tgbotapi.NewUpdate(0)
	upd.Timeout = 60

	//Получаем обновления от бота
	for update := range bot.GetUpdatesChan(upd) {

		ToID := t.GetID(update)
		if ToID == 0 {
			continue
		}
		srcUser := u.FindUser(&users, ToID, t.GetFrom(update).UserName)

		//Вывод данных о сообщении только для владельца
		if conf.Debug && (u.GetUser(srcUser).Status == u.SU) {
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
			if (update.Message.Text == key) && !conf.KeyUsed {
				userName := update.Message.From.UserName
				id := update.Message.From.ID
				conf.UseKey()
				idx := -1
				for userID, user := range users {
					if user.ID == id {
						idx = userID
						break
					}
				}

				if idx == -1 {
					users = append(users, *u.NewUser(id, u.SU, userName, "~"))
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
					h.Start(bot, srcUser)
				case "/main":
					h.Main(bot, srcUser)
				case "/help":
					h.Help(bot, srcUser)
				default:
					h.NoCmd(bot, srcUser)
				}
			} else {
				//Если просто текст
				msg := tgbotapi.NewMessage(ToID, "Not text-command TODO")
				bot.Send(msg)
			}

			//Проверка типа на событие кнопки
		} else if update.CallbackQuery != nil {
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			parts := strings.Split(update.CallbackQuery.Data, ".")

			//Вывод данных об ошибке только для владельца
			_, err := bot.Request(callback)
			if conf.Debug && (err != nil) && (u.GetUser(srcUser).Status == u.SU) {
				msg := tgbotapi.NewMessage(
					ToID, fmt.Sprintf("Callback error: \n%s", err),
				)
				bot.Send(msg)
				continue
			}

			//Проверка события
			switch parts[0] {
			case "user":
				h.UserInfo(bot, srcUser, &parts)
			case "users":
				h.UserList(bot, srcUser)
			case "select":
				h.SelectStatus(bot, srcUser, &parts)
			case "set":
				h.SetStatus(bot, srcUser, &parts)
			case "transferq":
				h.Transferq(bot, srcUser, &parts)
			case "transfer":
				h.Transfer(bot, srcUser, &parts)
			case "config":
				h.SetDebug(bot, &conf.Debug, srcUser, &parts) //TODO Config
			case "power", "powerq":
				h.RequestPower(bot, srcUser)
			case "terminal", "files":
				h.TODO(bot, srcUser)
			case "help":
				h.Help(bot, srcUser)
			case "debug":
				h.SetDebug(bot, &conf.Debug, srcUser, &parts)
			default:
				h.NoCmd(bot, srcUser)
			}
		}
	}
}
