package main

import (
	"BaumanS2SBot/internal/application"
	"BaumanS2SBot/internal/infrastructure/storage/cache"
	"BaumanS2SBot/internal/model"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strconv"
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
	StateConfirmationRequestForHelp
	StateSendingRequestForHelp
	StateDeletingRequestForHelp
)

func initDB(dataSourceName string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func maintainDBConnection(dataSourceName string, db *sqlx.DB) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := db.Ping(); err != nil {
					log.Println("Lost database connection. Reconnecting...")
					err = db.Close()
					if err != nil {
						log.Printf("Closing db %v", err)
					}
					db, err = initDB(dataSourceName)
					if err != nil {
						log.Fatalf("Error re-establishing connection to database: %v", err)
					}
				}
			}
		}
	}()
}

func getCategories(ctx context.Context, db *sqlx.DB) map[int]string {
	categories, err := application.GetCategoriesMap(ctx, db)
	if err != nil {
		log.Fatalf("can't take categories map %v", err)
	}
	return categories
}

func main() {
	var originMessage tgbotapi.CopyMessageConfig
	var categoryChosen string
	var helpCategoryID int
	var parsedDateTime time.Time
	var dateTimeText string
	var originMessageID int
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Fatal(err)
	}

	userStates := make(map[int64]int)
	token := os.Getenv("TELEGRAM_API_TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Error with the token: %v\n", err)
	}
	dataSourceName := os.Getenv("DATASOURCE_NAME")
	db, err := initDB(dataSourceName)
	//db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	maintainDBConnection(dataSourceName, db)

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

	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			_, messageIDToDelete := cache.DeleteExpiredRequests(
				"./internal/infrastructure/storage/cache/cache.json", loc)
			if len(messageIDToDelete) != 0 {
				log.Print(messageIDToDelete)
				for chatID, messageID := range messageIDToDelete {
					for _, DeleteID := range messageID {
						log.Print(DeleteID)
						msg := tgbotapi.NewDeleteMessage(chatID, DeleteID)
						if _, err = bot.Send(msg); err != nil {
							log.Printf("Error deleting expired messages: %v", err)
						}
					}
				}
			}
		}
	}()

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.CallbackQuery != nil {
			callbackData := update.CallbackQuery.Data
			log.Print(callbackData)
			var originMessageIDStr string
			parts := strings.SplitN(callbackData, ":", 2)
			log.Print(callbackData)
			if len(parts) == 2 {
				// Вторая часть будет содержать originMessageID
				originMessageIDStr = parts[1]

				// Преобразование строки в int
				originMessageID, err = strconv.Atoi(originMessageIDStr)
				if err != nil {
					fmt.Println("Ошибка при преобразовании:", err)
					return
				}

				log.Println("Origin Message ID:", originMessageID)
				log.Print(parts)
			} else {
				log.Println("Строка не содержит ожидаемый формат")
			}
			if parts[0] == "deleteRequest" {
				_, deleteMap := cache.DeleteRequest("./internal/infrastructure/storage/cache/cache.json", originMessageID)
				if len(deleteMap) != 0 {
					for chatID, messageID := range deleteMap {
						for _, DeleteID := range messageID {
							log.Print(DeleteID)
							msg := tgbotapi.NewDeleteMessage(chatID, DeleteID)
							if _, err = bot.Send(msg); err != nil {
								log.Printf("Error deleting expired messages: %v", err)
							}
						}
						callbackConfig := tgbotapi.NewCallback(update.CallbackQuery.ID, "Ура!!! 🎉")
						if _, err := bot.Request(callbackConfig); err != nil {
							log.Print(err)
						}
					}
				}
				edit := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					"Вам помогли с этим запросом 🎉")
				if _, err := bot.Send(edit); err != nil {
					log.Printf("Error editing msg: %v", err)
				}

			}
		}
		if update.Message == nil {
			continue
		}
		userID := update.Message.From.ID

		if isNewUser(db, userID) && update.Message.Text == "/start" {
			application.SendRegisterKeyboard(bot, update.Message.Chat.ID)
			userStates[userID] = StateStart
		}
		log.Printf("%v", userStates[userID])

		switch userStates[userID] {
		case StateStart:
			if update.Message.Text == "Зарегистрироваться" {
				ctx, cancel := context.WithTimeout(ctx, 10*time.Second)

				user := model.User{
					UserId:   userID,
					Username: update.Message.From.UserName,
				}
				if err = application.RegisterUser(ctx, db, user); err != nil {
					log.Printf("Error registering user: %v", err)
					cancel()
					continue
				}
				cancel()
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы успешно зарегистрированы!")
				if _, err = bot.Send(msg); err != nil {
					log.Printf("Error with register user:%v", err)
				}
				application.SendHomeKeyboard(bot, update.Message.Chat.ID, userStates, userID, StateHome)

			}
		case StateHome:
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
			} // TODO:
			// else if update.Message.Text == "Удалить или отредактировать запросы на помощь" {
			//	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Теперь вы можете удалять запросы на помощь")
			//	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			//		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Вернуться на главный экран")))
			//	if _, err := bot.Send(msg); err != nil {
			//		log.Printf("Error with sending u can delete keyboard: %v", err)
			//	}
			//	userStates[userID] = StateDeletingRequestForHelp
			// }
		case StateDeletingRequestForHelp:
			log.Print(StateDeletingRequestForHelp)
			if update.Message.Text == "Вернуться на главный экран" {
				application.SendHomeKeyboard(bot, update.Message.Chat.ID, userStates, userID, StateHome)
			}

		case StateChoosingCategoryForHelp:
			if update.Message.Text == "Вернуться на главный экран" {
				application.SendHomeKeyboard(bot, update.Message.Chat.ID, userStates, userID, StateHome)
			} else if categoryID, found := getKeyByValue(getCategories(ctx, db), update.Message.Text); found {
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
				application.SendHomeKeyboard(bot, update.Message.Chat.ID, userStates, userID, StateHome)
			} else {
				// date
				dateTimeText = update.Message.Text

				parsedDateTime, err = time.ParseInLocation(dateTimeLayout, dateTimeText, loc)

				if err != nil {
					log.Printf("Error while parsing date and time: %v", err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неправильный формат\n Введите в формате ЧЧ:ММ Д.М.Г (Пример: 19:15 01.12.2023)")
					if _, err := bot.Send(msg); err != nil {
						log.Printf("Error with sending error msg  %v", err)
					}
					userStates[userID] = StateFormingRequestForHelp
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите описание проблемы:")
					if _, err := bot.Send(msg); err != nil {
						log.Fatalf("Error with sending Введите описание msg: %v", err)
					}
					userStates[userID] = StateConfirmationRequestForHelp
				}
			}
		case StateConfirmationRequestForHelp:
			originMessageID = update.Message.MessageID
			originMessage = tgbotapi.NewCopyMessage(update.Message.Chat.ID,
				update.Message.Chat.ID, originMessageID)
			if update.Message.Text == "Вернуться на главный экран" {
				application.SendHomeKeyboard(bot, update.Message.Chat.ID, userStates, userID, StateHome)
			} else {
				application.SendConfirmationKeyboard(bot, update.Message.Chat.ID)
				userStates[userID] = StateSendingRequestForHelp
			}

		case StateSendingRequestForHelp:
			if update.Message.Text == "Вернуться на главный экран" {
				application.SendHomeKeyboard(bot, update.Message.Chat.ID, userStates, userID, StateHome)
			} else if update.Message.Text == "Да" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					fmt.Sprintf("Вы успешно сформировали запрос на помощь по теме: <b>%v</b>\nОписание: \n", categoryChosen))
				msg.ParseMode = "HTML"
				if _, err := bot.Send(msg); err != nil {
					log.Fatalf("Can't send congrats forming request: %v", err)
				}
				inlineBtn := tgbotapi.NewInlineKeyboardButtonData("Мне помогли ! 🎉", fmt.Sprintf("deleteRequest:%v", originMessageID))
				inlineKbd := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(inlineBtn))

				originMessage.ReplyMarkup = inlineKbd

				if _, err := bot.Send(originMessage); err != nil {
					log.Fatalf("Can't send cograts forming request: %v", err)
				}
				cleverUserIDSlice, err := application.GetCleverUsersSlice(ctx, db, helpCategoryID)
				if err != nil {
					log.Fatalf("can't get clever user's id %v", err)
				}
				log.Print(update.Message.MessageID)

				for _, cleverUserID := range cleverUserIDSlice {

					// кто отправил и дедлайн

					msg := tgbotapi.NewMessage(cleverUserID, fmt.Sprintf("Тема <b>%v</b> \n"+
						"Отправил пользователь с id: @%v "+
						"\nАктульно до %v\nОписание:",
						categoryChosen,
						update.Message.From.UserName,
						dateTimeText))
					msg.ParseMode = "HTML"

					sentMsg, err := bot.Send(msg)
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

					// описание задачи
					forwardMsg := tgbotapi.NewCopyMessage(cleverUserID,
						update.Message.Chat.ID, originMessageID)

					sentDescriptionMsg, err := bot.Send(forwardMsg)
					if err != nil {
						log.Fatalf("Can't forward message to clever guys with id: %v %v", cleverUserID, err)
					}

					err = cache.AddRequest("./internal/infrastructure/storage/cache/cache.json",
						cleverUserID, originMessageID, parsedDateTime, sentDescriptionMsg.MessageID)
					log.Print(cleverUserID, originMessageID, parsedDateTime, sentDescriptionMsg.MessageID)
					if err != nil {
						log.Fatalf("can't addRequest to json file: %v", err)
					}

				}

				application.SendHomeKeyboard(bot, update.Message.Chat.ID, userStates, userID, StateHome)
			} else if update.Message.Text == "Нет" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите описание проблемы:")
				if _, err := bot.Send(msg); err != nil {
					log.Fatalf("Error with sending Введите описание msg: %v", err)
				}
				userStates[userID] = StateConfirmationRequestForHelp
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нажмите на кнопку или введите: Да/Нет")
				if _, err := bot.Send(msg); err != nil {
					log.Fatalf("Error with sending Введите описание msg: %v", err)
				}
				userStates[userID] = StateSendingRequestForHelp
			}

		case StateAddCategory:
			var m string
			if categoryId, found := getKeyByValue(getCategories(ctx, db), update.Message.Text); found {
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
				application.SendHomeKeyboard(bot, update.Message.Chat.ID, userStates, userID, StateHome)
			} else if update.Message.Text == "Удалить категории" {
				application.SendUserRemoveCategoriesKeyboard(ctx, bot, db, update.Message.Chat.ID,
					application.GetCurrentUserCategoriesKeyboard(ctx, db, update.Message.Chat.ID))

				userStates[userID] = StateRemoveCategory
			}

		case StateRemoveCategory:
			if categoryId, found := getKeyByValue(getCategories(ctx, db), update.Message.Text); found {
				err := application.RemoveCategories(ctx, db, userID, categoryId)

				if err != nil {
					log.Printf("Error removing category: %v", err)
					continue
				}

				application.SendUserRemoveCategoriesKeyboard(ctx, bot, db, update.Message.Chat.ID,
					application.GetCurrentUserCategoriesKeyboard(ctx, db, update.Message.Chat.ID))
			} else if update.Message.Text == "Вернуться на главный экран" {
				application.SendHomeKeyboard(bot, update.Message.Chat.ID, userStates, userID, StateHome)
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

func isNewUser(db *sqlx.DB, userID int64) bool {
	var count int
	err := db.Get(&count, "SELECT count(*) FROM users WHERE user_id = $1", userID)
	if err != nil {
		log.Printf("Error querying user: %v", err)
		return false
	}
	return count == 0
}
