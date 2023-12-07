package application

import (
	"BaumanS2SBot/internal/application/states"
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"log"
)

func Page(update tgbotapi.Update, ctx context.Context, db *sqlx.DB,
	bot *tgbotapi.BotAPI, userID int64, userStates map[int64]int, chatID int64) {
	SendHomeKeyboard(bot, update.Message.Chat.ID, userStates, userID, states.StateHome)
	if update.Message.Text == "Хочу помогать" {
		SendCategorySelectKeyboard(ctx, db, chatID, update, bot)

		userStates[userID] = states.StateAddCategory
	} else if update.Message.Text == "Нужна помощь" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите предмет:")
		msg.ReplyMarkup = GetAllCategoryKeyboard(ctx, db)

		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending need help message %v", err)
		}

		userStates[userID] = states.StateChoosingCategoryForHelp
	}
}
