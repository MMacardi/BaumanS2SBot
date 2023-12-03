package application

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"log"
	"strings"
)

func AddKeyboardButton(keyboard tgbotapi.ReplyKeyboardMarkup, newButton string) tgbotapi.ReplyKeyboardMarkup {

	button := tgbotapi.NewKeyboardButton(newButton)
	maxButtonsPerRow := 3
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

func GetAllCategoryKeyboard(ctx context.Context, db *sqlx.DB) tgbotapi.ReplyKeyboardMarkup {
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

func GetCategorySelectKeyboard(ctx context.Context, db *sqlx.DB) tgbotapi.ReplyKeyboardMarkup {
	categories, err := GetCategoriesMap(ctx, db)

	if err != nil {
		log.Fatalf("Error getting categories %v", err)
	}

	CategoriesKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Вернуться на главный экран"), tgbotapi.NewKeyboardButton("Удалить категории")))
	for _, categoryName := range categories {
		CategoriesKeyboard = AddKeyboardButton(CategoriesKeyboard, categoryName)
	}

	return CategoriesKeyboard

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

	msg := tgbotapi.NewMessage(chatID, "Вы зарегистрированы в категориях"+" "+currentCategoriesString)

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
