package application

import (
	"BaumanS2SBot/internal/application/commands"
	"BaumanS2SBot/internal/application/media"
	"BaumanS2SBot/internal/application/states"
	"BaumanS2SBot/internal/infrastructure/storage/cache"
	"BaumanS2SBot/internal/model"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"log"
	"strconv"
	"strings"
	"time"
)

func IsNewUser(db *sqlx.DB, userID int64) bool {
	var count int
	err := db.Get(&count, "SELECT count(*) FROM users WHERE user_id = $1", userID)
	if err != nil {
		log.Printf("Error querying user: %v", err)
		return false
	}
	return count == 0
}

func InitDB(dataSourceName string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func MaintainDBConnection(dataSourceName string, db *sqlx.DB, ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			if err := db.Ping(); err != nil {
				log.Println("Lost database connection. Reconnecting...")
				err = db.Close()
				if err != nil {
					log.Printf("Closing db %v", err)
				}
				db, err = InitDB(dataSourceName)
				if err != nil {
					log.Fatalf("Error re-establishing connection to database: %v", err)
				}
			}
		}
	}
}

func DeleteExpiredRequests(bot *tgbotapi.BotAPI, loc *time.Location, ticker *time.Ticker) {
	for range ticker.C {
		_, messageIDToDelete, messageIDToEdit := cache.DeleteExpiredRequestsFromCache(
			"./internal/infrastructure/storage/cache/cache.json", loc)
		if len(messageIDToDelete) != 0 {
			log.Print(messageIDToDelete)
			for chatID, messageID := range messageIDToDelete {
				for _, deleteID := range messageID {
					log.Print(deleteID)
					msg := tgbotapi.NewDeleteMessage(chatID, deleteID)
					if _, err := bot.Send(msg); err != nil {
						log.Printf("Error deleting expired messages: %v", err)
					}
				}
			}
			for chatID, messageID := range messageIDToEdit {
				for _, editID := range messageID {
					msg := tgbotapi.NewMessage(chatID, "У запроса с данным ниже описанием прошел срок, он удален\n"+
						"Если помощь с ним еще нужна сформируйте его еще раз")
					if _, err := bot.Send(msg); err != nil {
						log.Printf("Error deleting expired messages: %v", err)
					}
					msgFwd := tgbotapi.NewCopyMessage(chatID, chatID, editID)
					if _, err := bot.Send(msgFwd); err != nil {
						log.Printf("Error deleting expired messages: %v", err)
					}
					msg1 := tgbotapi.NewDeleteMessage(chatID, editID)
					if _, err := bot.Send(msg1); err != nil {
						log.Printf("Error deleting expired messages: %v", err)
					}

				}
			}
		}
	}
}

func Start(userSessions map[int64]*model.UserSession, update tgbotapi.Update, ctx context.Context, db *sqlx.DB,
	bot *tgbotapi.BotAPI, userID int64, chatID int64, currentState int, userStates map[int64]int, dateTimeLayout string,
	loc *time.Location, debug bool) {

	if _, ok := userSessions[userID]; !ok {
		userSessions[userID] = &model.UserSession{}
	}

	session := userSessions[userID]

	AddCommandsMenu(bot)

	if update.Message.IsCommand() {
		switch update.Message.Command() {
		case "start":
			if IsNewUser(db, userID) {
				SendRegisterKeyboard(bot, update.Message.Chat.ID)
				userStates[userID] = states.StateStart
			} else {
				SendHomeKeyboard(bot, chatID, userStates, userID, states.StateHome)
			}
		case "help":
			commands.SendHelpMessage(bot, chatID)
		}
	}

	switch currentState {
	case states.StateHome:
		Page(update, ctx, db, bot, userID, userStates, chatID)
	case states.StateStart:
		User(update, ctx, db, bot, userID, userStates)
	case states.StateAddCategory:
		Add(update, ctx, db, bot, userID, userStates, chatID)
	case states.StateRemoveCategory:
		Remove(update, ctx, db, bot, userID, userStates)
	case states.StateChoosingCategoryForHelp:
		ChooseCategory(session, update, ctx, db, bot, userID, userStates)
	case states.StateFormingRequestForHelp:
		FormingRequest(session, update, bot, userID, userStates, dateTimeLayout, loc)
	case states.StateConfirmationRequestForHelp:
		ConfirmRequest(session, update, bot, userID, chatID, userStates)
	case states.StateSendingRequestForHelp:
		SendingRequest(session, ctx, db, update, bot, userID, userStates, debug)
	case states.StateUserRequestsForHelp:
		// TODO
	}

	return
}

func ProcessCallback(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	callbackData := update.CallbackQuery.Data
	log.Print(callbackData)
	parts := strings.SplitN(callbackData, ":", 2)
	if len(parts) != 2 {
		log.Println("Строка не содержит ожидаемый формат")
		return
	}
	originMessageIDStr := parts[1]
	originMessageID, err := strconv.Atoi(originMessageIDStr)
	if err != nil {
		fmt.Println("Ошибка при преобразовании:", err)
		return
	}
	switch parts[0] {

	case "deleteRequest":
		DeleteCallback(update, bot, originMessageID)

	}
}

func DeleteCallback(update tgbotapi.Update, bot *tgbotapi.BotAPI, originMessageID int) {
	callbackConfig := tgbotapi.NewCallback(update.CallbackQuery.ID, "Ура!!! 🎉")
	if _, err := bot.Request(callbackConfig); err != nil {
		log.Print(err)
	}

	originMessage := update.CallbackQuery.Message
	deleteMap, editMap, _ := cache.DeleteRequest("./internal/infrastructure/storage/cache/cache.json", originMessageID)
	if len(deleteMap) == 0 {
		return
	}

	for chatID, innerMap := range editMap {
		for forwardMessageID, isMedia := range innerMap {
			if forwardMessageID == originMessage.MessageID {
				media.IfExist(isMedia, chatID, forwardMessageID, bot, "Вам помогли с этим запросом 🎉")
			} else {
				media.IfExist(isMedia, chatID, forwardMessageID, bot, "Ваш запрос был удален администратором")
			}
		}
	}

	for chatID, messageID := range deleteMap {
		for _, DeleteID := range messageID {
			msg := tgbotapi.NewDeleteMessage(chatID, DeleteID)
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Error deleting expired messages: %v", err)
			}

		}
	}

}
