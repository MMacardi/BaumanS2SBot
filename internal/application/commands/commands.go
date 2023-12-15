package commands

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

const (
	NeedHelp         = "Нужна помощь 🆘"
	WannaHelp        = "Хочу помогать 🤝"
	BackToHome       = "Вернуться на главный экран 🏠"
	RemoveCategories = "Удалить предметы 🗑️"
	Yes              = "Да 🤩"
	No               = "Нет 🤔"

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
