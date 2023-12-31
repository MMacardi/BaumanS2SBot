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
	"time"
)

func GetCleverUsersSlice(ctx context.Context, db *sqlx.DB, categoryID int) ([]int64, error) {
	query := `SELECT user_id FROM user_categories WHERE category_id = $1`

	var cleverUserIDSlice []int64

	err := db.SelectContext(ctx, &cleverUserIDSlice, query, categoryID)

	if err != nil {
		return nil, err
	}

	return cleverUserIDSlice, nil
}

func ChooseCategory(session *model.UserSession, update tgbotapi.Update, ctx context.Context, db *sqlx.DB,
	bot *tgbotapi.BotAPI, userID int64, userStates map[int64]int) {
	if update.Message.Text == commands.BackToHome {
		SendHomeKeyboard(bot, update.Message.Chat.ID, userStates, userID)
	} else if categoryID, found := GetKeyByValue(GetCategories(ctx, db), update.Message.Text); found {
		categoryChosen := update.Message.Text
		session.CategoryChosen = categoryChosen
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбран предмет: "+
			"<b>"+categoryChosen+"</b>"+
			"\nНапишите сроки вашего запроса на помощь в формате Часы:Минуты Дата.Месяц.Год (Пример: 19:15 01.12.2023)")
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(commands.BackToHome)))
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error with sending chosen category msg %v", err)
		}

		userStates[userID] = states.StateFormingRequestForHelp
		helpCategoryID := categoryID
		session.HelpCategoryID = helpCategoryID
		return
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите предмет из клавиатуры ниже:")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error with sending chosen category msg %v", err)
		}
		userStates[userID] = states.StateChoosingCategoryForHelp
	}
	return

}

func FormingRequest(session *model.UserSession, update tgbotapi.Update,
	bot *tgbotapi.BotAPI, userID int64, userStates map[int64]int, dateTimeLayout string, loc *time.Location) {
	if update.Message.Text == commands.BackToHome {
		SendHomeKeyboard(bot, update.Message.Chat.ID, userStates, userID)
		return
	} else {
		// date
		dateTimeText := update.Message.Text

		parsedDateTime, err := time.ParseInLocation(dateTimeLayout, dateTimeText, loc)

		userYear := parsedDateTime.Year()
		now := time.Now().In(loc)
		currentYear := now.Year()

		if err != nil || parsedDateTime.Before(now) || userYear > currentYear+1 {
			log.Printf("Error while parsing date and time: %v", err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неправильный формат "+
				"или введена некорректная дата (сроки указываются в пределах одного года)"+
				"\nВведите в формате ЧЧ:ММ Д.М.Г (Пример: 19:15 01.12.2023)")
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Error with sending error msg  %v", err)
			}
			userStates[userID] = states.StateFormingRequestForHelp
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите описание проблемы:")
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Error with sending Введите описание msg: %v", err)
			}
			userStates[userID] = states.StateConfirmationRequestForHelp
		}
		session.DateTimeText = dateTimeText
		session.ParsedDateTime = parsedDateTime
		return
	}
}

func ConfirmRequest(session *model.UserSession, update tgbotapi.Update,
	bot *tgbotapi.BotAPI, userID int64, chatID int64, userStates map[int64]int) {
	if update.Message.Text == commands.BackToHome {
		SendHomeKeyboard(bot, update.Message.Chat.ID, userStates, userID)
		return
	}
	if update.Message.Sticker != nil {
		msg := tgbotapi.NewMessage(chatID, "Некорректное описание (нельзя использовать стикер в качестве описания)")

		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending sticker msg: %v", err)
		}
		userStates[userID] = states.StateConfirmationRequestForHelp
		return
	}
	session.IsMedia = media.Exist(update.Message)
	originMessageID := update.Message.MessageID
	session.OriginMessageID = originMessageID
	originMessage := tgbotapi.NewCopyMessage(chatID,
		chatID, originMessageID)
	session.OriginMessage = originMessage
	SendConfirmationKeyboard(bot, update.Message.Chat.ID)
	userStates[userID] = states.StateSendingRequestForHelp
	return
}

func SendingRequest(session *model.UserSession, ctx context.Context, db *sqlx.DB, update tgbotapi.Update,
	bot *tgbotapi.BotAPI, userID int64, userStates map[int64]int, debug bool, fwdToSelf bool) {
	if update.Message.Text == commands.BackToHome {
		SendHomeKeyboard(bot, update.Message.Chat.ID, userStates, userID)
	} else if update.Message.Text == commands.Yes {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("Вы успешно сформировали запрос на помощь по теме: <b>%v</b>\nСрок до: <b>%v</b>\nОписание: \n", session.CategoryChosen, session.DateTimeText))
		msg.ParseMode = "HTML"
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("Can't send congrats forming request: %v", err)
		}

		inlineBtn := tgbotapi.NewInlineKeyboardButtonData("Мне помогли ! 🎉", fmt.Sprintf("deleteRequest:%v", session.OriginMessageID))
		inlineKbd := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(inlineBtn))

		session.OriginMessage.ReplyMarkup = inlineKbd

		origMsg, err := bot.Send(session.OriginMessage)
		if err != nil {
			log.Printf("Can't send cograts forming request: %v", err)
		}

		err = cache.AddRequest("./internal/infrastructure/storage/cache/cache.json",
			userID,
			0,
			session.ParsedDateTime,
			origMsg.MessageID,
			session.IsMedia)

		if err != nil {
			log.Printf("can't addRequest to json file: %v", err)
		}

		cleverUserIDSlice, err := GetCleverUsersSlice(ctx, db, session.HelpCategoryID)

		if err != nil {
			log.Printf("can't get clever user's id %v", err)
		}

		// admin method lmao
		var adminID int64 = 865277762
		if debug == false {
			cleverUserIDSlice = append(cleverUserIDSlice, adminID)
		}
		for _, cleverUserID := range cleverUserIDSlice {
			if cleverUserID != userID && fwdToSelf == false {
				SendingToCleverUsers(session, update, bot, session.IsMedia, cleverUserID)
			} else if fwdToSelf == true {
				SendingToCleverUsers(session, update, bot, session.IsMedia, cleverUserID)
			}
		}
		if debug == false {

			session.OriginMessage.ChatID = adminID

			if _, err = bot.Send(tgbotapi.NewMessage(adminID, "Админ лови, но учти - если ты создал запрос, то на верхнюю кнопочку,"+
				"после нажатия, ты ничего не удалишь, потому что ты удалил уже на нижнюю :0")); err != nil {
				log.Printf("Error Sending msg to admin %v", err)
			}

			msg = tgbotapi.NewMessage(adminID, fmt.Sprintf("#ЗапросНаПомощь\nТема: #<b>%v</b> \n"+
				"Отправил пользователь с id: @%v "+
				"\nАктуально до: %v\nОписание:",
				session.CategoryChosen,
				update.Message.From.UserName,
				session.DateTimeText))
			msg.ParseMode = "HTML"

			_, err = bot.Send(msg)
			if err != nil {
				log.Printf("Can't forward message to clever guys with id: %v %v", adminID, err)
			}

			if _, err = bot.Send(session.OriginMessage); err != nil {
				log.Printf("Can't send cograts forming request: %v", err)
			}

			_, err = bot.Send(tgbotapi.NewMessage(adminID, "Это конец запроса, админ"))

		}

		SendHomeKeyboard(bot, update.Message.Chat.ID, userStates, userID)
		return
	} else if update.Message.Text == commands.No {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите описание проблемы:")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error with sending Введите описание msg: %v", err)
		}
		userStates[userID] = states.StateConfirmationRequestForHelp
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нажмите на кнопку:")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error with sending Введите описание msg: %v", err)
		}
		userStates[userID] = states.StateSendingRequestForHelp
	}
	return
}

func SendingToCleverUsers(session *model.UserSession, update tgbotapi.Update,
	bot *tgbotapi.BotAPI, isMedia bool, cleverUserID int64) {
	// кто отправил и дедлайн

	msg := tgbotapi.NewMessage(cleverUserID, fmt.Sprintf("#ЗапросНаПомощь\nТема: #<b>%v</b> \n"+
		"Отправил пользователь с id: @%v "+
		"\nАктуально до: %v\nОписание:",
		session.CategoryChosen,
		update.Message.From.UserName,
		session.DateTimeText))
	msg.ParseMode = "HTML"

	sentMsg, err := bot.Send(msg)
	if err != nil {
		log.Printf("Can't forward message to clever guys with id: %v %v", cleverUserID, err)
	}

	err = cache.AddRequest("./internal/infrastructure/storage/cache/cache.json",
		cleverUserID,
		session.OriginMessageID,
		session.ParsedDateTime,
		sentMsg.MessageID,
		false)

	if err != nil {
		log.Printf("can't addRequest to json file: %v", err)
	}

	// описание задачи
	forwardMsg := tgbotapi.NewCopyMessage(cleverUserID,
		update.Message.Chat.ID, session.OriginMessageID)

	sentDescriptionMsg, err := bot.Send(forwardMsg)
	if err != nil {
		log.Printf("Can't forward message to clever guys with id: %v %v", cleverUserID, err)
	}

	err = cache.AddRequest("./internal/infrastructure/storage/cache/cache.json",
		cleverUserID, session.OriginMessageID, session.ParsedDateTime, sentDescriptionMsg.MessageID, isMedia)
	//log.Print(cleverUserID, session.OriginMessageID, session.ParsedDateTime, sentDescriptionMsg.MessageID)
	if err != nil {
		log.Printf("can't addRequest to json file: %v", err)
	}

}

// TODO: func SendUserRequests()
