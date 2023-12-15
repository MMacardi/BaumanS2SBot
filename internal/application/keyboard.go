package application

import (
	"BaumanS2SBot/internal/application/commands"
	"BaumanS2SBot/internal/application/states"
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"log"
	"sort"
	"strings"
)

func AddCommandsMenu(bot *tgbotapi.BotAPI) {
	commandsKeyboard := []tgbotapi.BotCommand{
		{Command: "start", Description: "Зарегистрироваться, либо перейти на главный экран"},
		{Command: "help", Description: "Узнать о работе бота"},
	}

	_, err := bot.Request(tgbotapi.NewSetMyCommands(commandsKeyboard...))
	if err != nil {
		log.Panic(err)
	}
}

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

func getSortedCategoriesSlice(ctx context.Context, db *sqlx.DB) []string {
	categories, err := GetCategoriesMap(ctx, db)
	if err != nil {
		log.Fatalf("Error getting categories %v", err)
	}

	var categoryNames []string
	for _, name := range categories {
		categoryNames = append(categoryNames, name)
	}
	sort.Strings(categoryNames)

	return categoryNames
}

func GetHelpCategoryKeyboard(ctx context.Context, db *sqlx.DB) tgbotapi.ReplyKeyboardMarkup {
	categories := getSortedCategoriesSlice(ctx, db)

	categoriesKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(commands.BackToHome)))

	for _, categoryName := range categories {
		categoriesKeyboard = AddKeyboardButton(categoriesKeyboard, categoryName)
	}

	return categoriesKeyboard

}

func GetCategorySelectKeyboard(ctx context.Context, db *sqlx.DB, chatID int64) tgbotapi.ReplyKeyboardMarkup {
	categoryNames := getSortedCategoriesSlice(ctx, db)

	categoriesKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(commands.BackToHome), tgbotapi.NewKeyboardButton(commands.RemoveCategories)))
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
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы зарегистрированы в предметах:"+
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
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(commands.BackToHome)))
	for _, currentCategoryName := range currentCategories {
		selectCategoriesKeyboard = AddKeyboardButton(selectCategoriesKeyboard, currentCategoryName)
	}

	return selectCategoriesKeyboard
}

func SendUserRemoveCategoriesMsgKeyboard(ctx context.Context, bot *tgbotapi.BotAPI, db *sqlx.DB, chatID int64, currentCategoriesKeyboard tgbotapi.ReplyKeyboardMarkup) {
	currentCategoriesString := strings.Join(GetCategoriesNameByCategoryID(ctx, db,
		GetUserCurrentCategoriesSlice(ctx, db, chatID)), ",")

	msg := tgbotapi.NewMessage(chatID, "Вы зарегистрированы в предметах:"+" "+currentCategoriesString)

	if currentCategoriesString == "" {
		msg = tgbotapi.NewMessage(chatID, "Вам нечего удалять :(")
	}

	msg.ReplyMarkup = currentCategoriesKeyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending send home message: %v", err)
	}
}

func SendHomeKeyboard(bot *tgbotapi.BotAPI, chatID int64, userStates map[int64]int, userID int64) {
	msg := tgbotapi.NewMessage(chatID, "Что вы хотите сделать?")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(commands.NeedHelp),
			tgbotapi.NewKeyboardButton(commands.WannaHelp),
			// tgbotapi.NewKeyboardButton("Удалить или отредактировать запросы на помощь"),
		),
	)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending send home message: %v", err)
	}
	userStates[userID] = states.StateHome
}

func SendRegisterKeyboard(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Нажмите, чтобы начать:")
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
			tgbotapi.NewKeyboardButton(commands.Yes),
			tgbotapi.NewKeyboardButton(commands.No),
			tgbotapi.NewKeyboardButton(commands.BackToHome),
		),
	)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending welcome message: %v", err)
	}
}
