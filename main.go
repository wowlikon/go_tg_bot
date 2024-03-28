package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var debug, key_used bool
var Owner user

type Access int

const (
	Unregistered Access = iota // EnumIndex = 0
	Waiting                    // EnumIndex = 1
	Member                     // EnumIndex = 2
	Admin                      // EnumIndex = 3
	SU                         // EnumIndex = 4
)

func (w Access) String() string {
	return [...]string{"Unregistered", "Waiting", "Member", "Admin", "SU"}[w]
}

func (w Access) EnumIndex() int {
	return int(w)
}

type user struct {
	ID        int64  `json:"id"`
	Status    Access `json:"status"`
	UserName  string `json:"name"`
	Directory string `json:"directory"`
}

func generateKey(n int) (string, error) {
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

func main() {
	var users []user

	//Проверка аргументов запуска
	args := os.Args
	fmt.Println(args)
	if (GetIndex(args, "-h") != -1) || (GetIndex(args, "--help") != -1) {
		fmt.Printf("Usage: %s [arguments]\n", args[0])
		fmt.Println("\t-h --help  | help information")
		fmt.Println("\t-d --debug | enable debug info")
		return
	}

	debug = (GetIndex(args, "-d") != -1) || (GetIndex(args, "--debug") != -1)

	//Загружаем .env
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}

	key_len, _ := strconv.Atoi(os.Getenv("KEY_LENGTH"))
	key, _ := generateKey(key_len)
	fmt.Printf("Admin key: %s\n", key)

	//Создаем бота
	fmt.Println("Starting bot")
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Bot @%s is online", bot.Self.UserName)
	if debug {
		fmt.Println(" in debug mode")
	} else {
		fmt.Println()
	}

	//Устанавливаем время обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	//Получаем обновления от бота
	for update := range bot.GetUpdatesChan(u) {

		//Вывод данных о сообщении
		if debug {
			var ToID int64
			if update.Message != nil {
				ToID = update.Message.Chat.ID
			}
			if update.CallbackQuery != nil {
				ToID = update.CallbackQuery.Message.Chat.ID
			}
			if ToID == 0 {
				continue
			}
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
					users = append(users, user{id, SU, userName, "~"})
				} else {
					users[idx].Status = SU
				}

				//Приветствие для суперадминистратора
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("```welcome $sudo hello_world --admin %s```", userName))
				msg.ParseMode = "MarkdownV2"
				bot.Send(msg)
				continue
			}

			//Проверка команды
			if strings.HasPrefix(update.Message.Text, "/") {
				parts := strings.Split(update.Message.Text, " ")
				switch parts[0] {
				case "/start":
					start(bot, update, &users)
				case "/users":
					userList(bot, update, &users)
				case "/status":
					status(bot, update, &users)
				case "/debug":
					toggleDebug(bot, update, &users, parts)
				case "/help":
					help(bot, update)
				default:
					nocmd(bot, update)
				}
			} else {
				//Если просто текст
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "not command TODO")
				bot.Send(msg)
			}

			//Проверка типа на событие кнопки
		} else if update.CallbackQuery != nil {
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf("callback %s", update.CallbackQuery.Data))
			bot.Send(msg)
		}
	}
}

func start(bot *tgbotapi.BotAPI, update tgbotapi.Update, users *[]user) {
	var msg tgbotapi.MessageConfig

	//Проверка на повторный запуск
	userName := update.Message.From.UserName
	id := update.Message.From.ID
	alreadyExists := false
	for _, user := range *users {
		if user.ID == id {
			alreadyExists = true
			break
		}
	}

	//Приветствие для пользователя
	if !alreadyExists {
		*users = append(*users, user{id, Waiting, userName, "~"})
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Hello, %s", userName))
	} else {
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Already exist")
	}
	bot.Send(msg)
}

func userList(bot *tgbotapi.BotAPI, update tgbotapi.Update, users *[]user) {
	var msg tgbotapi.MessageConfig

	//Команда только для админов
	allow := false
	for _, user := range *users {
		if user.ID == update.Message.Chat.ID {
			allow = user.Status >= Admin
			break
		}
	}

	if !allow {
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Access denied")
		bot.Send(msg)
		return
	}

	//Вывод списка пользователей
	if debug {
		usersJSON, err := json.MarshalIndent(
			*users, "", "  ",
		)
		if err != nil {
			msg := tgbotapi.NewMessage(
				update.Message.Chat.ID,
				fmt.Sprintf(
					"Error marshaling users to JSON: \n%s", err,
				),
			)
			bot.Send(msg)
		}
		msg = tgbotapi.NewMessage(
			update.Message.Chat.ID, fmt.Sprintf("```json\n%s\n```", usersJSON),
		)
	} else {
		//Добавление кнопок для перехода
		ikb := tgbotapi.NewInlineKeyboardMarkup()
		kb := make([][]tgbotapi.InlineKeyboardButton, 0, len(*users))

		for _, user := range *users {
			txt := fmt.Sprintf("%s (%s)", user.UserName, user.Status)
			ikbRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonURL(txt, fmt.Sprintf("tg://openmessage?user_id=%d", user.ID)))
			kb = append(kb, ikbRow)
		}

		exRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Example", "hello.world"))
		kb = append(kb, exRow)

		msg = tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Here are the users:",
		)
		ikb.InlineKeyboard = kb
		msg.ReplyMarkup = ikb
	}
	msg.ParseMode = "MarkdownV2"
	bot.Send(msg)
}

func toggleDebug(bot *tgbotapi.BotAPI, update tgbotapi.Update, users *[]user, parts []string) {
	my_status := Unregistered
	for _, user := range *users {
		if user.ID == update.Message.Chat.ID {
			my_status = user.Status
			break
		}
	}
	if my_status != SU {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Access denied!")
		bot.Send(msg)
		return
	}
	if len(parts) == 1 {
		parts = append(parts, "")
	}
	if strings.ToLower(parts[1]) == "on" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Debug mode on!")
		bot.Send(msg)
		debug = true
		return
	}
	if strings.ToLower(parts[1]) == "off" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Debug mode off!")
		bot.Send(msg)
		debug = false
		return
	}
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		fmt.Sprintf("Debug: %t \n/debug [on/off]", debug),
	)
	bot.Send(msg)
}

func status(bot *tgbotapi.BotAPI, update tgbotapi.Update, users *[]user) {
	my_status := Unregistered
	for _, user := range *users {
		if user.ID == update.Message.Chat.ID {
			my_status = user.Status
			break
		}
	}
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		fmt.Sprintf("You're status: %s", my_status),
	)
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
