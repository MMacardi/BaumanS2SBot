package main

import (
	"BaumanS2SBot/internal/model"
	"BaumanS2SBot/internal/storage"
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strings"
	"time"
)

const (
	StateHome = iota
	StateStart
	StateAddCategory
	StateRemoveCategory
)

func main() {
	userStates := make(map[int64]int)
	token := goDotEnvVariable("TELEGRAM_APITOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Error with the token: %v\n", err)
	}
	dataSourceName := goDotEnvVariable("DATASOURCENAME")
	db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	defer db.Close()

	log.Println("Bot has been started...")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	ctx := context.Background()
	categories, err := storage.GetCategoriesMap(ctx, db)

	categoriesKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Вернуться на главный экран"),
			tgbotapi.NewKeyboardButton("Удалить категории")))

	for _, categoryName := range categories {
		categoriesKeyboard = storage.AddKeyboardButton(categoriesKeyboard, categoryName)
	}
	if err != nil {
		log.Fatalf("Error getting categories", err)
	}

	for update := range updates {
		userID := update.Message.From.ID
		currentState := userStates[userID]
		if update.Message == nil { // если сообщение отсутствует, пропускаем итерацию
			continue
		}
		//_, er := storage.GetEveryUserCategoriesMap(ctx, db)
		//if er != nil {
		//	log.Fatal("Error with db", er)
		//}
		if isNewUser(db, userID) {
			userStates[userID] = StateStart
			log.Print(userStates)
		}

		switch currentState {
		case StateStart:
			sendRegisterKeyboard(bot, update.Message.Chat.ID)

			if update.Message.Text == "Зарегистрироваться" {
				ctx, cancel := context.WithTimeout(ctx, 10*time.Second)

				user := model.User{
					UserId:   userID,
					Username: update.Message.From.UserName,
				}
				if err := storage.RegisterUser(ctx, db, user); err != nil {
					log.Printf("Error registering user: %v %s", err)
					cancel()
					continue
				}
				cancel()
				sendHomeKeyboard(bot, update.Message.Chat.ID)
				userStates[userID] = StateHome
			}
		case StateHome:
			sendHomeKeyboard(bot, update.Message.Chat.ID)
			if update.Message.Text == "Хочу помогать" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы зарегистрированы в категориях"+" "+strings.Join(storage.GetCategoriesNameByCategoryID(ctx, db,
					storage.GetUserCategoriesSlice(ctx, db, update.Message.Chat.ID)), ","))

				msg.ReplyMarkup = categoriesKeyboard
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Error sending registration confirmation message: %v", err)
				}

				userStates[userID] = StateAddCategory
			}

		case StateAddCategory:
			if categoryId, found := getKeyByValue(categories, update.Message.Text); found {
				err := storage.AddCategories(ctx, db, userID, categoryId)
				if err != nil {
					log.Printf("Error sending category message: %v", err)
					continue
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы зарегистрированы в категориях"+" "+strings.Join(storage.GetCategoriesNameByCategoryID(ctx, db,
					storage.GetUserCategoriesSlice(ctx, db, update.Message.Chat.ID)), ","))
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Error sending category response: %v", err)
					continue
				}
			} else if update.Message.Text == "Вернуться на главный экран" {
				sendHomeKeyboard(bot, update.Message.Chat.ID)
				userStates[userID] = StateHome
			} else if update.Message.Text == "Удалить категории" {
				sendUserCategoriesKeyboard(ctx, bot, db, update.Message.Chat.ID, getCurrentUserCategories(ctx, db, update.Message.Chat.ID))

				userStates[userID] = StateRemoveCategory
			}
		case StateRemoveCategory:
			if categoryId, found := getKeyByValue(categories, update.Message.Text); found {
				err := storage.RemoveCategories(ctx, db, userID, categoryId)

				if err != nil {
					log.Printf("Error removing category: %v", err)
					continue
				}

				sendUserCategoriesKeyboard(ctx, bot, db, update.Message.Chat.ID, getCurrentUserCategories(ctx, db, update.Message.Chat.ID))
			} else if update.Message.Text == "Вернуться на главный экран" {
				sendHomeKeyboard(bot, update.Message.Chat.ID)
				userStates[userID] = StateHome
			}

		}
	}
}

func getCurrentUserCategories(ctx context.Context, db *sqlx.DB, chatID int64) tgbotapi.ReplyKeyboardMarkup {
	currentCategories := storage.GetCategoriesNameByCategoryID(ctx, db, storage.GetUserCategoriesSlice(ctx, db, chatID))
	currentCategoriesKeyboard := tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Вернуться на главный экран")))

	for _, currentCategoryName := range currentCategories {
		currentCategoriesKeyboard = storage.AddKeyboardButton(currentCategoriesKeyboard, currentCategoryName)
	}

	return currentCategoriesKeyboard
}

func sendUserCategoriesKeyboard(ctx context.Context, bot *tgbotapi.BotAPI, db *sqlx.DB, chatID int64, currentCategoriesKeyboard tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, "Вы зарегистрированы в категориях"+" "+strings.Join(storage.GetCategoriesNameByCategoryID(ctx, db,
		storage.GetUserCategoriesSlice(ctx, db, chatID)), ","))

	msg.ReplyMarkup = currentCategoriesKeyboard

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending send home message: %v", err)
	}
}

func sendHomeKeyboard(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Что вы хотите сделать?")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Нужна помощь"),
			tgbotapi.NewKeyboardButton("Хочу помогать"),
		),
	)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending send home message: %v", err)
	}
}
func sendRegisterKeyboard(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Добро пожаловать! Нажмите для регистрации:")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Зарегистрироваться"),
		),
	)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending welcome message: %v", err)
	}
}

func getKeyByValue(myMap map[int]string, valueToFind string) (int, bool) {
	for key, value := range myMap {
		if value == valueToFind {
			return key, true // Возвращаем ключ и true, если нашли совпадение
		}
	}
	return 0, false // Возвращаем 0 и false, если совпадение не найдено
}

func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func isNewUser(db *sqlx.DB, userID int64) bool {
	var count int
	err := db.Get(&count, "SELECT count(*) FROM users WHERE user_id = $1", userID)
	if err != nil {
		log.Printf("Error querying user: %v", err)
		return false
	}
	return count == 0
}

//switch update.Message.Text {
//case "open":
//msg.ReplyMarkup = numericKeyboard
//case "1", "2", "3", "4", "5", "6":
//msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
//default:
//responseText = "Please, choose from the list below"

//if update.CallbackQuery != nil {
//	// TODO: забивка инфы в датабазу
//	// ID сообщения и чата, чтобы изменить сообщение
//	chatID := update.CallbackQuery.Message.Chat.ID
//	messageID := update.CallbackQuery.Message.MessageID
//
//	// Новый текст сообщения и удаление клавиатуры
//	newMsg := tgbotapi.NewEditMessageText(chatID, messageID, "Изменено")
//	newMsg.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{
//		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
//	}
//
//	// Отправляем изменение сообщения
//	_, err := bot.Send(newMsg)
//	if err != nil {
//		// Обработка ошибки
//		log.Println("can't send", err)
//
