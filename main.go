package main

import (
	"TelegramS2SBot/internal/model"
	"TelegramS2SBot/internal/storage"
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"os"
	"time"
)

var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("1.com", "http://1.com"),
		tgbotapi.NewInlineKeyboardButtonData("2", "2"),
		tgbotapi.NewInlineKeyboardButtonData("3", "3"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("4", "4"),
		tgbotapi.NewInlineKeyboardButtonData("5", "5"),
		tgbotapi.NewInlineKeyboardButtonData("6", "6"),
	),
)

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
	err := db.Get(&count, "SELECT count(*) FROM users WHERE userid = $1", userID)
	if err != nil {
		log.Printf("Error querying user: %v", err)
		return false
	}
	return count == 0
}

func main() {
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

	for update := range updates {
		if update.Message == nil { // если сообщение отсутствует, пропускаем итерацию
			continue
		}

		userID := update.Message.From.ID
		// Регистрация
		// Проверяем, является ли пользователь новым, прежде чем отправлять приветственное сообщение
		if update.Message.Text == "/start" && isNewUser(db, userID) {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добро пожаловать! Нажмите для регистрации:")
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Зарегистрироваться"),
				),
			)
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Error sending welcome message: %v", err)
			}
			continue
		}

		// Обработка сообщения "Зарегистрироваться"
		if update.Message.Text == "Зарегистрироваться" && isNewUser(db, userID) {
			// Тут должна быть логика для регистрации пользователя, чтобы он больше не считался новым
			// ...

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите категорию:")
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Матан"),
					tgbotapi.NewKeyboardButton("Инжа"),
				),
			)
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Error sending category selection message: %v", err)
			}
			continue
		}

		// Обработка выбора категории
		if update.Message.Text == "Матан" || update.Message.Text == "Инжа" {
			// Предполагаем, что функция storage.RegisterUser регистрирует пользователя и обновляет его статус
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			category := update.Message.Text
			// Обязательно обработайте возможную ошибку от RegisterUser
			user := model.User{
				UserId:   userID,
				Name:     update.Message.From.UserName,
				Category: category,
			}
			if err := storage.RegisterUser(ctx, db, user); err != nil {
				log.Printf("Error registering user: %v %s", err)
				cancel()
				continue
			}
			cancel() // Убедитесь, что контекст закрыт в случае успешной регистрации

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы зарегистрированы в категории: "+category)
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Error sending registration confirmation message: %v", err)
			}
		}
	}
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
