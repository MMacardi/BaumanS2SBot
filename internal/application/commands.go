package application

import (
	"BaumanS2SBot/internal/application/states"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"log"
)

const (
	NeedHelpCmd         = "Нужна помощь 🆘"
	WannaHelpCmd        = "Хочу помогать 🤝"
	BackToHomeCmd       = "Вернуться на главный экран 🏠"
	RemoveCategoriesCmd = "Удалить предметы 🗑️"
	YesCmd              = "Да 🤩"
	NoCmd               = "Нет 🤔"

	Help = `
Для того чтобы помогать другим пользователям, нажмите на главном экране "Хочу помогать 🤝" и выберите предмет
Для того чтобы перестать помогать в определённом предмете нажмите на кнопку "Удалить предметы 🗑️"

Для того чтобы получить помощь составьте запрос на кнопку "Нужна помощь 🆘" и далее следуя инструкциям

<b> !!! Учтите, что для оказания вам помощи, у вас должно быть свое имя пользователя телеграм, указать его вы можете,` +
		` перейдя в самом Телеграме по пути: Настройки - Мой аккаунт - Имя пользователя </b>
`
)

func SendHelpMessage(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, Help)
	msg.ParseMode = "HTML"
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending help msg %v", err)
	}
}

func CommandHandler(update tgbotapi.Update, bot *tgbotapi.BotAPI, db *sqlx.DB, userID int64, chatID int64, userStates map[int64]int) {
	switch update.Message.Command() {
	case "start":
		if IsNewUser(db, userID) {
			SendRegisterKeyboard(bot, update.Message.Chat.ID)
			userStates[userID] = states.StateStart
		} else {
			SendHomeKeyboard(bot, chatID, userStates, userID)
		}
	case "help":
		SendHelpMessage(bot, chatID)
	}
}
