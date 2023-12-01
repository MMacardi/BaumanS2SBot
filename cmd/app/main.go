package main

import (
	"BaumanS2SBot/internal/application"
	"BaumanS2SBot/internal/infrastructure/storage/cache"
	"BaumanS2SBot/internal/model"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strings"
	"time"
)

const dateTimeLayout = "15:04 02.01.2006"

const (
	StateHome = iota
	StateStart
	StateAddCategory
	StateRemoveCategory
	StateChoosingCategoryForHelp
	StateFormingRequestForHelp
	StateSendingRequestForHelp
)

func main() {
	var originMessage tgbotapi.CopyMessageConfig
	var categoryChosen string
	var helpCategoryID int
	var parsedDateTime time.Time
	var dateTimeText string
	userStates := make(map[int64]int)
	token := goDotEnvVariable("TELEGRAM_API_TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Error with the token: %v\n", err)
	}
	dataSourceName := goDotEnvVariable("DATASOURCE_NAME")
	db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	defer func(db *sqlx.DB) {
		_ = db.Close()
	}(db)

	log.Println("Bot has been started...")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	ctx := context.TODO()
	categories, err := application.GetCategoriesMap(ctx, db)

	if err != nil {
		log.Fatalf("can't take categories map %v", err)
	}

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		userID := update.Message.From.ID
		currentState := userStates[userID]
		if update.Message == nil {
			continue
		}

		if isNewUser(db, userID) {
			userStates[userID] = StateStart
			log.Print(userStates)
		}
		log.Printf("%v", currentState)

		ticker := time.NewTicker(1 * time.Second)
		go func() {
			for range ticker.C {
				_, messageIDToDelete := cache.DeleteExpiredRequests(
					"./internal/infrastructure/storage/cache/cache.json")
				if len(messageIDToDelete) != 0 {
					for chatID, messageID := range messageIDToDelete {
						msg := tgbotapi.NewDeleteMessage(chatID, messageID)
						if _, err := bot.Send(msg); err != nil {
							log.Printf("Error deleting expired messages: %v", err)
						}
					}
				}
			}
		}()

		switch currentState {
		case StateStart:
			application.SendRegisterKeyboard(bot, update.Message.Chat.ID)

			if update.Message.Text == "Зарегистрироваться" {
				ctx, cancel := context.WithTimeout(ctx, 10*time.Second)

				user := model.User{
					UserId:   userID,
					Username: update.Message.From.UserName,
				}
				if err := application.RegisterUser(ctx, db, user); err != nil {
					log.Printf("Error registering user: %v", err)
					cancel()
					continue
				}
				cancel()
				application.SendHomeKeyboard(bot, update.Message.Chat.ID)
				userStates[userID] = StateHome
			}
		case StateHome:
			application.SendHomeKeyboard(bot, update.Message.Chat.ID)
			if update.Message.Text == "Хочу помогать" {
				categoriesString := application.GetCurrentUserCategoriesString(ctx, db, update.Message.Chat.ID)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы зарегистрированы в категориях"+
					" "+categoriesString)

				if categoriesString == "" {
					msg = tgbotapi.NewMessage(update.Message.Chat.ID,
						"Вы не зарегистрированы ни в одной из категорий :(")
				}

				msg.ReplyMarkup = application.GetCategorySelectKeyboard(ctx, db)
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Error sending registration confirmation message: %v", err)
				}

				userStates[userID] = StateAddCategory
			} else if update.Message.Text == "Нужна помощь" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите предмет:")
				msg.ReplyMarkup = application.GetAllCategoryKeyboard(ctx, db)

				if _, err := bot.Send(msg); err != nil {
					log.Printf("Error sending need help message %v", err)
				}

				userStates[userID] = StateChoosingCategoryForHelp
			}
		case StateChoosingCategoryForHelp:
			if update.Message.Text == "Вернуться на главный экран" {
				application.SendHomeKeyboard(bot, update.Message.Chat.ID)
				userStates[userID] = StateHome
			} else if categoryID, found := getKeyByValue(categories, update.Message.Text); found {
				categoryChosen = update.Message.Text
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбрана категория: "+
					"<b>"+categoryChosen+"</b>"+
					"\nНапишите дедлайн вашего запроса на помощь в формате ЧЧ:ММ Д.М.Г (Пример: 19:15 01.12.2023)")
				msg.ParseMode = "HTML"
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Вернуться на главный экран")))
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Error with sending chosen category msg %v", err)
				}

				userStates[userID] = StateFormingRequestForHelp
				helpCategoryID = categoryID
			}
		case StateFormingRequestForHelp:
			if update.Message.Text == "Вернуться на главный экран" {
				application.SendHomeKeyboard(bot, update.Message.Chat.ID)
				userStates[userID] = StateHome
			} else {
				// date
				dateTimeText = update.Message.Text

				parsedDateTime, err = time.Parse(dateTimeLayout, dateTimeText)

				if err != nil {
					log.Printf("Error while parsing date and time: %v", err)
				}
				log.Print(parsedDateTime)
				userStates[userID] = StateSendingRequestForHelp

			}
		case StateSendingRequestForHelp:
			if update.Message.Text == "Вернуться на главный экран" {
				application.SendHomeKeyboard(bot, update.Message.Chat.ID)
				userStates[userID] = StateHome
			} else {
				cleverUserIDSlice, err := application.GetCleverUsersSlice(ctx, db, helpCategoryID)
				if err != nil {
					log.Fatalf("can't get clever user's id %v", err)
				}
				originMessageID := update.Message.MessageID
				originMessage = tgbotapi.NewCopyMessage(update.Message.Chat.ID,
					update.Message.Chat.ID, originMessageID)
				for _, cleverUserID := range cleverUserIDSlice {
					forwardMsg := tgbotapi.NewCopyMessage(cleverUserID,
						update.Message.Chat.ID, originMessageID)

					msg := tgbotapi.NewMessage(cleverUserID, fmt.Sprintf("Тема <b>%v</b> \n"+
						"Отправил пользователь с id: @%v "+
						"\nАктульно до %v",
						categoryChosen,
						update.Message.From.UserName,
						dateTimeText))
					sentMsg, err := bot.Send(forwardMsg)
					if err != nil {
						log.Fatalf("Can't forward message to clever guys with id: %v %v", cleverUserID, err)
					}

					err = cache.AddRequest("./internal/infrastructure/storage/cache/cache.json",
						cleverUserID, originMessageID, parsedDateTime, sentMsg.MessageID)

					if err != nil {
						log.Fatalf("can't addRequest to json file: %v", err)
					}

					sentMsg, err = bot.Send(msg)
					if err != nil {
						log.Fatalf("Can't forward message to clever guys with id: %v %v", cleverUserID, err)
					}

					err = cache.AddRequest("./internal/infrastructure/storage/cache/cache.json",
						cleverUserID,
						originMessageID,
						parsedDateTime,
						sentMsg.MessageID)

					if err != nil {
						log.Fatalf("can't addRequest to json file: %v", err)
					}

				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"Вы успешно сформировали запрос на помощь по теме: <b>"+
						categoryChosen+
						"</b>\nОписание: \n")
				msg.ParseMode = "HTML"

				if _, err := bot.Send(msg); err != nil {
					log.Fatalf("Can't send cograts forming request: %v", err)
				}

				if _, err := bot.Send(originMessage); err != nil {
					log.Fatalf("Can't send cograts forming request: %v", err)
				}

				application.SendHomeKeyboard(bot, update.Message.Chat.ID)
				userStates[userID] = StateHome
			}

		case StateAddCategory:
			var m string
			if categoryId, found := getKeyByValue(categories, update.Message.Text); found {
				if application.IsCategoryAdded(ctx, db, update.Message.Chat.ID, update.Message.Text) {
					m = "Вы уже зарегистрированы в категории: " + update.Message.Text + "\n\n"
				} else {
					err := application.AddCategories(ctx, db, userID, categoryId)
					if err != nil {
						log.Printf("Error sending category message: %v", err)
						continue
					}
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, m+
					"Вы зарегистрированы в категориях: "+
					strings.Join(application.GetCategoriesNameByCategoryID(ctx, db,
						application.GetUserCurrentCategoriesSlice(ctx, db, update.Message.Chat.ID)), ","))
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Error sending category response: %v", err)
					continue
				}

			} else if update.Message.Text == "Вернуться на главный экран" {
				application.SendHomeKeyboard(bot, update.Message.Chat.ID)
				userStates[userID] = StateHome
			} else if update.Message.Text == "Удалить категории" {
				application.SendUserRemoveCategoriesKeyboard(ctx, bot, db, update.Message.Chat.ID,
					application.GetCurrentUserCategoriesKeyboard(ctx, db, update.Message.Chat.ID))

				userStates[userID] = StateRemoveCategory
			}

		case StateRemoveCategory:
			if categoryId, found := getKeyByValue(categories, update.Message.Text); found {
				err := application.RemoveCategories(ctx, db, userID, categoryId)

				if err != nil {
					log.Printf("Error removing category: %v", err)
					continue
				}

				application.SendUserRemoveCategoriesKeyboard(ctx, bot, db, update.Message.Chat.ID,
					application.GetCurrentUserCategoriesKeyboard(ctx, db, update.Message.Chat.ID))
			} else if update.Message.Text == "Вернуться на главный экран" {
				application.SendHomeKeyboard(bot, update.Message.Chat.ID)
				userStates[userID] = StateHome
			}

		}
	}
}

func getKeyByValue(myMap map[int]string, valueToFind string) (int, bool) {
	for key, value := range myMap {
		if value == valueToFind {
			return key, true
		}
	}
	return 0, false
}

func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load("cmd/app/.env")

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
