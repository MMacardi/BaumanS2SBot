package application

import (
	"BaumanS2SBot/internal/application/states"
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"log"
	"strings"
)

func Remove(update tgbotapi.Update, ctx context.Context, db *sqlx.DB,
	bot *tgbotapi.BotAPI, userID int64, userStates map[int64]int) {
	if categoryId, found := GetKeyByValue(GetCategories(ctx, db), update.Message.Text); found {
		err := RemoveCategories(ctx, db, userID, categoryId)

		if err != nil {
			log.Printf("Error removing category: %v", err)
			return
		}

		SendUserRemoveCategoriesKeyboard(ctx, bot, db, update.Message.Chat.ID,
			GetCurrentUserCategoriesKeyboard(ctx, db, update.Message.Chat.ID))
	} else if update.Message.Text == "Вернуться на главный экран" {
		SendHomeKeyboard(bot, update.Message.Chat.ID, userStates, userID, states.StateHome)
	}
}

func Add(update tgbotapi.Update, ctx context.Context, db *sqlx.DB,
	bot *tgbotapi.BotAPI, userID int64, userStates map[int64]int, chatID int64) {
	var m string
	categoryName := update.Message.Text[:len(update.Message.Text)-4]
	if categoryId, found := GetKeyByValue(GetCategories(ctx, db), categoryName); found {
		if IsCategoryAdded(ctx, db, update.Message.Chat.ID, categoryName) {
			m = "Вы уже зарегистрированы в предмете: " + categoryName + "\n\n"
		} else {
			err := AddCategories(ctx, db, userID, categoryId)
			if err != nil {
				log.Printf("Error adding user category: %v", err)
				return
			}
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, m+
			"Вы зарегистрированы в предметах: "+
			strings.Join(GetCategoriesNameByCategoryID(ctx, db,
				GetUserCurrentCategoriesSlice(ctx, db, update.Message.Chat.ID)), ","))
		msg.ReplyMarkup = GetCategorySelectKeyboard(ctx, db, chatID)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending category response: %v", err)
			return
		}

	} else if update.Message.Text == "Вернуться на главный экран" {
		SendHomeKeyboard(bot, update.Message.Chat.ID, userStates, userID, states.StateHome)
	} else if update.Message.Text == "Удалить предметы" {
		SendUserRemoveCategoriesKeyboard(ctx, bot, db, update.Message.Chat.ID,
			GetCurrentUserCategoriesKeyboard(ctx, db, update.Message.Chat.ID))

		userStates[userID] = states.StateRemoveCategory
	}

}
