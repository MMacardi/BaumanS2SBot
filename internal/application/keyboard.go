package application

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"log"
	"sort"
	"strings"
)

func AddKeyboardButton(keyboard tgbotapi.ReplyKeyboardMarkup, newButton string) tgbotapi.ReplyKeyboardMarkup {

	button := tgbotapi.NewKeyboardButton(newButton)
	maxButtonsPerRow := 2
	// if no buttons
	if len(keyboard.Keyboard) == 0 {
		keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(button))
	} else {

		lastRowIndex := len(keyboard.Keyboard) - 1
		if len(keyboard.Keyboard[lastRowIndex]) < maxButtonsPerRow {
			// add button to the last row
			keyboard.Keyboard[lastRowIndex] = append(keyboard.Keyboard[lastRowIndex], button)
		} else {
			// if row is full create new row with button
			keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(button))
		}
	}

	return keyboard
}

func GetHelpCategoryKeyboard(ctx context.Context, db *sqlx.DB) tgbotapi.ReplyKeyboardMarkup {
	categories, err := GetCategoriesMap(ctx, db)

	if err != nil {
		log.Fatalf("Error getting categories %v", err)
	}

	CategoriesKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Вернуться на главный экран")))
	for _, categoryName := range categories {
		CategoriesKeyboard = AddKeyboardButton(CategoriesKeyboard, categoryName)
	}

	return CategoriesKeyboard

}

func GetCategorySelectKeyboard(ctx context.Context, db *sqlx.DB, chatID int64) tgbotapi.ReplyKeyboardMarkup {
	categories, err := GetCategoriesMap(ctx, db)
	if err != nil {
		log.Fatalf("Error getting categories %v", err)
	}

	var categoryNames []string
	for _, name := range categories {
		categoryNames = append(categoryNames, name)
	}
	sort.Strings(categoryNames)

	categoriesKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Вернуться на главный экран"), tgbotapi.NewKeyboardButton("Удалить предметы")))
	tick := ""
	userCategories := GetCurrentUserCategories(ctx, db, chatID)
	for _, categoryName := range categoryNames {
		tick = "❌"
		for _, item := range userCategories {
			if categoryName == item {
				tick = "✅"
				break
			}
		}
		categoriesKeyboard = AddKeyboardButton(categoriesKeyboard, categoryName+" "+tick)
	}
	return categoriesKeyboard
}

func SendCategorySelectKeyboard(ctx context.Context, db *sqlx.DB, chatID int64, update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	categoriesString := GetCurrentUserCategoriesString(ctx, db, update.Message.Chat.ID)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы зарегистрированы в предметах"+
		" "+categoriesString)

	if categoriesString == "" {
		msg = tgbotapi.NewMessage(update.Message.Chat.ID,
			"Вы не зарегистрированы ни в одном предмете :(")
	}
	msg.ReplyMarkup = GetCategorySelectKeyboard(ctx, db, chatID)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending registration confirmation message: %v", err)
	}
}

func GetCurrentUserCategoriesKeyboard(ctx context.Context, db *sqlx.DB, chatID int64) tgbotapi.ReplyKeyboardMarkup {
	currentCategories := GetCategoriesNameByCategoryID(ctx, db, GetUserCurrentCategoriesSlice(ctx, db, chatID))
	selectCategoriesKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Вернуться на главный экран")))
	for _, currentCategoryName := range currentCategories {
		selectCategoriesKeyboard = AddKeyboardButton(selectCategoriesKeyboard, currentCategoryName)
	}

	return selectCategoriesKeyboard
}

func SendUserRemoveCategoriesKeyboard(ctx context.Context, bot *tgbotapi.BotAPI, db *sqlx.DB, chatID int64, currentCategoriesKeyboard tgbotapi.ReplyKeyboardMarkup) {
	currentCategoriesString := strings.Join(GetCategoriesNameByCategoryID(ctx, db,
		GetUserCurrentCategoriesSlice(ctx, db, chatID)), ",")

	msg := tgbotapi.NewMessage(chatID, "Вы зарегистрированы в предметах"+" "+currentCategoriesString)

	if currentCategoriesString == "" {
		msg = tgbotapi.NewMessage(chatID, "Вам нечего удалять :(")
	}

	msg.ReplyMarkup = currentCategoriesKeyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending send home message: %v", err)
	}
}

func SendHomeKeyboard(bot *tgbotapi.BotAPI, chatID int64, userStates map[int64]int, userID int64, StateHome int) {
	msg := tgbotapi.NewMessage(chatID, "Что вы хотите сделать?")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Нужна помощь"),
			tgbotapi.NewKeyboardButton("Хочу помогать"),
			// tgbotapi.NewKeyboardButton("Удалить или отредактировать запросы на помощь"),
		),
	)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending send home message: %v", err)
	}
	userStates[userID] = StateHome
}

func SendRegisterKeyboard(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Добро пожаловать! Нажмите, чтобы начать:")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Начать"),
		),
	)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending welcome message: %v", err)
	}
}

func SendConfirmationKeyboard(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Вы уверены в правильности запроса?")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Да"),
			tgbotapi.NewKeyboardButton("Нет"),
		),
	)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending welcome message: %v", err)
	}
}
