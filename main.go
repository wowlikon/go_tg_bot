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

type Access int

const (
	Unregistered Access = iota // EnumIndex = 0
	NoAccess                   // EnumIndex = 1
	Waiting                    // EnumIndex = 2
	Member                     // EnumIndex = 3
	Admin                      // EnumIndex = 4
	SU                         // EnumIndex = 5
)

func (w Access) String() string {
	return [...]string{"Unregistered", "NoAccess", "Waiting", "Member", "Admin", "SU"}[w]
}

func (w Access) EnumIndex() int {
	return int(w)
}

func AccessList() []Access {
	return []Access{0, 1, 2, 3, 4, 5}
}

type user struct {
	ID          int64  `json:"id"`
	Status      Access `json:"status"`
	UserName    string `json:"name"`
	Directory   string `json:"directory"`
	EditMessage int64  `json:"edit_msg"`
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

func FindUser(users *[]user, id int64) user {
	for _, user := range *users {
		if user.ID == id {
			return user
		}
	}
	return user{id, Unregistered, "Unknown", "~", 0}
}

func FindID(update tgbotapi.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.Message.Chat.ID
	}
	return 0
}

func USend(bot tgbotapi.BotAPI, emsg *tgbotapi.EditMessageTextConfig) {
	_, err := bot.Send(*emsg)
	if err != nil {
		msg := tgbotapi.NewMessage(emsg.ChatID, emsg.Text)
		msg.ReplyMarkup = emsg.ReplyMarkup
		bot.Send(msg)
	}
}

func main() {
	var users []user

	//Проверка аргументов запуска
	args := os.Args
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

	fmt.Printf("Bot @%s is online in ", bot.Self.UserName)
	if debug {
		fmt.Println("debug mode")
	} else {
		fmt.Println("standart mode")
	}

	//Устанавливаем время обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	//Получаем обновления от бота
	for update := range bot.GetUpdatesChan(u) {

		ToID := FindID(update)
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
					users = append(users, user{id, SU, userName, "~", 0})
				} else {
					users[idx].Status = SU
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
				userInfo(bot, update, &users, parts)
			case "users":
				userList(bot, update, &users)
			case "select":
				selectStatus(bot, update, &users, parts)
			case "set":
				setStatus(bot, update, &users, parts)
			case "transferq":
				transferq(bot, update, &users, parts)
			case "transfer":
				transfer(bot, update, &users, parts)
			default:
				nocmd(bot, update)
			}
		}
	}
}

// Функции команд
func start(bot *tgbotapi.BotAPI, update tgbotapi.Update, users *[]user) {
	var msg tgbotapi.MessageConfig
	ToID := FindID(update)

	//Проверка на повторный запуск
	userName := update.Message.From.UserName
	id := update.Message.From.ID
	exist := (FindUser(users, id).Status != Unregistered)

	//Приветствие для пользователя
	if !exist {
		*users = append(*users, user{id, Waiting, userName, "~", 0})
		msg = tgbotapi.NewMessage(ToID, fmt.Sprintf("Hello, %s", userName))
	} else {
		msg = tgbotapi.NewMessage(ToID, "Already exist")
	}
	bot.Send(msg)
}

func userList(bot *tgbotapi.BotAPI, update tgbotapi.Update, users *[]user) {
	var msg tgbotapi.MessageConfig

	//Команда только для админов
	ToID := FindID(update)
	if ToID == 0 {
		return
	}
	if FindUser(users, ToID).Status < Admin {
		msg := tgbotapi.NewMessage(ToID, "Access denied")
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
				ToID,
				fmt.Sprintf(
					"Error marshaling users to JSON: \n%s", err,
				),
			)
			bot.Send(msg)
		}
		msg = tgbotapi.NewMessage(
			ToID, fmt.Sprintf("```json\n%s\n```", usersJSON),
		)
	} else {
		//Добавление кнопок для перехода
		ikb := tgbotapi.NewInlineKeyboardMarkup()
		kb := make([][]tgbotapi.InlineKeyboardButton, 0, len(*users))

		for _, user := range *users {
			txt := fmt.Sprintf("%s (%s)", user.UserName, user.Status)
			//ikbRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonURL(txt, fmt.Sprintf("tg://openmessage?user_id=%d", user.ID)))
			ikbRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(txt, fmt.Sprintf("user.%d", user.ID)))
			kb = append(kb, ikbRow)
		}

		msg = tgbotapi.NewMessage(
			ToID,
			"Here are the users:",
		)
		ikb.InlineKeyboard = kb
		msg.ReplyMarkup = ikb
	}
	msg.ParseMode = "MarkdownV2"
	bot.Send(msg)
}

func toggleDebug(bot *tgbotapi.BotAPI, update tgbotapi.Update, users *[]user, parts []string) {
	ToID := FindID(update)
	my_status := FindUser(users, ToID).Status
	if my_status != SU {
		msg := tgbotapi.NewMessage(ToID, "Access denied!")
		bot.Send(msg)
		return
	}

	if len(parts) == 1 {
		parts = append(parts, "")
	}

	if strings.ToLower(parts[1]) == "on" {
		msg := tgbotapi.NewMessage(ToID, "Debug mode on!")
		bot.Send(msg)
		debug = true
		return
	}

	if strings.ToLower(parts[1]) == "off" {
		msg := tgbotapi.NewMessage(ToID, "Debug mode off!")
		bot.Send(msg)
		debug = false
		return
	}

	msg := tgbotapi.NewMessage(
		ToID,
		fmt.Sprintf("Debug: %t \n/debug [on/off]", debug),
	)
	bot.Send(msg)
}

func status(bot *tgbotapi.BotAPI, update tgbotapi.Update, users *[]user) {
	ToID := FindID(update)
	my_status := FindUser(users, ToID).Status
	msg := tgbotapi.NewMessage(
		ToID,
		fmt.Sprintf("You're status: %s", my_status),
	)
	bot.Send(msg)
}

func help(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(FindID(update), "TODO")
	bot.Send(msg)
}

func nocmd(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(FindID(update), "Error 404 command not found :(")
	bot.Send(msg)
}

// Функции событий
func userInfo(bot *tgbotapi.BotAPI, update tgbotapi.Update, users *[]user, parts []string) {
	ToID := FindID(update)
	other_id, err := strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		if debug {
			msg := tgbotapi.NewMessage(
				ToID,
				fmt.Sprintf("Callback parse ID error: \n %s", err),
			)
			bot.Send(msg)
		}
		return
	}

	other := FindUser(users, other_id)
	me := FindUser(users, ToID)

	if me.Status < Admin {
		msg := tgbotapi.NewMessage(ToID, "Permission denied")
		bot.Send(msg)
		return
	}

	//Текстовая информация
	msg := tgbotapi.NewMessage(
		ToID,
		fmt.Sprintf("Username: %s\nStatus: %s", other.UserName, other.Status),
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
	msg.ReplyMarkup = ikb
	bot.Send(msg)
}

func selectStatus(bot *tgbotapi.BotAPI, update tgbotapi.Update, users *[]user, parts []string) {
	ToID := FindID(update)
	other_id, err := strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		if debug {
			msg := tgbotapi.NewMessage(
				ToID,
				fmt.Sprintf("Callback parse ID error: \n %s", err),
			)
			bot.Send(msg)
		}
		return
	}

	other := FindUser(users, other_id)
	me := FindUser(users, ToID)

	//Добавление клавиш управления
	ikb := tgbotapi.NewInlineKeyboardMarkup()
	kb := make([][]tgbotapi.InlineKeyboardButton, 0, len(*users))

	if me.Status < Admin {
		msg := tgbotapi.NewMessage(ToID, "Permission denied")
		bot.Send(msg)
		return
	}

	for _, v := range AccessList() {
		if v == 0 {
			continue
		}

		if v != SU {
			ikbRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(v.String(), fmt.Sprintf("set.%s.%d", parts[1], v)))
			kb = append(kb, ikbRow)
		}
	}

	if me.Status == SU {
		ikbRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Transfer SU", fmt.Sprintf("transferq.%s", parts[1])))
		kb = append(kb, ikbRow)
	}

	msg := tgbotapi.NewMessage(
		ToID,
		fmt.Sprintf("Select %s's access level:", other.UserName),
	)
	ikb.InlineKeyboard = kb
	msg.ReplyMarkup = ikb
	bot.Send(msg)
}

func setStatus(bot *tgbotapi.BotAPI, update tgbotapi.Update, users *[]user, parts []string) {
	ToID := FindID(update)
	other_id, err := strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		if debug {
			msg := tgbotapi.NewMessage(
				ToID,
				fmt.Sprintf("Callback parse ID error: \n %s", err),
			)
			bot.Send(msg)
		}
		return
	}

	status_id, err := strconv.Atoi(parts[2])
	if err != nil {
		if debug {
			msg := tgbotapi.NewMessage(
				ToID,
				fmt.Sprintf("Callback parse ID error: \n %s", err),
			)
			bot.Send(msg)
		}
		return
	}

	if status_id == 0 {
		msg := tgbotapi.NewMessage(ToID, "Zero status error")
		bot.Send(msg)
		return
	}

	me := FindUser(users, ToID)
	if me.Status < Admin {
		msg := tgbotapi.NewMessage(ToID, "Permission denied")
		bot.Send(msg)
		return
	}

	name := ""
	for i, user := range *users {
		if user.ID == other_id {
			if (*users)[i].Status >= me.Status {
				msg := tgbotapi.NewMessage(ToID, "Permission denied")
				bot.Send(msg)
				return
			}
			name = user.UserName
			(*users)[i].Status = Access(status_id)
			break
		}
	}

	msg := tgbotapi.NewMessage(
		ToID,
		fmt.Sprintf("%s now %s", name, AccessList()[status_id]),
	)
	bot.Send(msg)
}

func transferq(bot *tgbotapi.BotAPI, update tgbotapi.Update, users *[]user, parts []string) {
	ToID := FindID(update)
	me := FindUser(users, ToID)
	other_id, err := strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		if debug {
			msg := tgbotapi.NewMessage(
				ToID,
				fmt.Sprintf("Callback parse ID error: \n %s", err),
			)
			bot.Send(msg)
		}
		return
	}

	if me.Status != SU {
		msg := tgbotapi.NewMessage(ToID, "Permission denied")
		bot.Send(msg)
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

	ikbRow = tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Yes", fmt.Sprintf("transfer.%s", parts[1])))
	kb = append(kb, ikbRow)

	ikbRow = tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("No", "users"))
	kb = append(kb, ikbRow)

	msg := tgbotapi.NewMessage(
		ToID,
		fmt.Sprintf(
			"Do you want to transfer super user access to %s\n(You lost own access and become administator)",
			name,
		),
	)

	ikb.InlineKeyboard = kb
	msg.ReplyMarkup = ikb
	bot.Send(msg)
}

func transfer(bot *tgbotapi.BotAPI, update tgbotapi.Update, users *[]user, parts []string) {
	ToID := FindID(update)
	me := FindUser(users, ToID)
	other_id, err := strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		if debug {
			msg := tgbotapi.NewMessage(
				ToID,
				fmt.Sprintf("Callback parse ID error: \n %s", err),
			)
			bot.Send(msg)
		}
		return
	}

	if me.Status != SU {
		msg := tgbotapi.NewMessage(ToID, "Permission denied")
		bot.Send(msg)
		return
	}

	new_su := ""
	for _, user := range *users {
		if user.ID == other_id {
			new_su = me.UserName
			me.Status = Admin
			user.Status = SU
			break
		}
	}

	if new_su != "" {
		msg := tgbotapi.NewMessage(
			ToID,
			fmt.Sprintf("%s is SU\nNow you administrator", new_su),
		)
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(ToID, "Error")
		bot.Send(msg)
	}
}

func noevent(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(FindID(update), "Error 404 callback not found :(")
	bot.Send(msg)
}
