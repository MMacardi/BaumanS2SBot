package media

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func Exist(originMessage *tgbotapi.Message) bool {
	if originMessage.Document != nil || (originMessage.Photo != nil && len(originMessage.Photo) > 0) || originMessage.Audio != nil || originMessage.Video != nil {
		return true
	}
	return false
}
func IfExist(isMedia bool, chatID int64, messageID int, bot *tgbotapi.BotAPI, msgText string) {
	if isMedia {
		edit := tgbotapi.NewEditMessageCaption(chatID,
			messageID,
			msgText)

		edit.ReplyMarkup = nil

		if _, err := bot.Send(edit); err != nil {
			log.Printf("Error editing msg with document: %v", err)
		}
		return
	}

	edit := tgbotapi.NewEditMessageText(chatID,
		messageID,
		msgText)

	edit.ReplyMarkup = nil

	if _, err := bot.Send(edit); err != nil {
		log.Printf("Error editing text msg: %v", err)
	}

}
