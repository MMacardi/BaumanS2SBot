package main

import (
	"BaumanS2SBot/internal/application"
	"BaumanS2SBot/internal/model"
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"os"
	"time"
)

const dateTimeLayout = "15:04 02.01.2006"

// 1.(идея цель проблема) технологии проблемы решение проблем архитектура
// показ рабочего проекта монетизация
// как привлекать умных людей

func main() {

	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Fatal(err)
	}

	token := os.Getenv("TELEGRAM_API_TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Error with the token: %v\n", err)
	}
	dataSourceName := os.Getenv("DATASOURCE_NAME")
	db, err := application.InitDB(dataSourceName)

	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	defer func(db *sqlx.DB) {
		_ = db.Close()
	}(db)

	log.Println("Bot has been started...")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	ctx := context.TODO()

	dbTicker := time.NewTicker(60 * time.Second)
	deleteTicker := time.NewTicker(5 * time.Second)

	go application.MaintainDBConnection(dataSourceName, db, dbTicker)
	go application.DeleteExpiredRequests(bot, loc, deleteTicker)

	var userStates = make(map[int64]int)
	var userSessions = make(map[int64]*model.UserSession)

	debug := false

	updates := bot.GetUpdatesChan(u)
	for update := range updates {

		if update.CallbackQuery != nil {
			application.ProcessCallback(update, bot)
		}
		if update.Message == nil {
			continue
		}

		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		application.Start(userSessions, update, ctx, db, bot, userID, chatID, userStates[userID], userStates, dateTimeLayout, loc, debug)

	}
}
