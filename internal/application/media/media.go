package media

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func Exist(originMessage *tgbotapi.Message, bot *tgbotapi.BotAPI, msgText string) {

	if originMessage.Document != nil {
		edit := tgbotapi.NewEditMessageCaption(originMessage.Chat.ID,
			originMessage.MessageID,
			msgText)

		if _, err := bot.Send(edit); err != nil {
			log.Printf("Error editing msg with document: %v", err)
		}
	} else if originMessage.Photo != nil && len(originMessage.Photo) > 0 {
		edit := tgbotapi.NewEditMessageCaption(originMessage.Chat.ID,
			originMessage.MessageID,
			msgText)

		if _, err := bot.Send(edit); err != nil {
			log.Printf("Error editing photo msg: %v", err)
		}
	} else if originMessage.Audio != nil {
		edit := tgbotapi.NewEditMessageCaption(originMessage.Chat.ID,
			originMessage.MessageID,
			msgText)

		if _, err := bot.Send(edit); err != nil {
			log.Printf("Error editing audio msg: %v", err)
		}
	} else if originMessage.Video != nil {
		edit := tgbotapi.NewEditMessageCaption(originMessage.Chat.ID,
			originMessage.MessageID,
			msgText)

		if _, err := bot.Send(edit); err != nil {
			log.Printf("Error editing video msg: %v", err)
		}
	} else {
		edit := tgbotapi.NewEditMessageText(originMessage.Chat.ID,
			originMessage.MessageID,
			msgText)

		if _, err := bot.Send(edit); err != nil {
			log.Printf("Error editing text msg: %v", err)
		}
	}
}
