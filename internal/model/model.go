package model

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

type User struct {
	Id       int    `db:"id"`
	UserId   int64  `db:"user_id"`
	Username string `db:"username"`
	Ranking  int    `db:"ranking"`
}

type Category struct {
	Id           int    `db:"id"`
	CategoryName string `db:"category_name"`
}

type FileData struct {
	ChatID           int64     `json:"chat_id"`
	OrigMessageID    int       `json:"orig_message_id"`
	ExpiryDate       time.Time `json:"expiry_date"`
	ForwardMessageID int       `json:"forward_message_id"`
	IsMedia          bool      `json:"is_media"`
}

type UserSession struct {
	OriginMessage   tgbotapi.CopyMessageConfig
	CategoryChosen  string
	HelpCategoryID  int
	ParsedDateTime  time.Time
	DateTimeText    string
	OriginMessageID int
	ChatID          int64
	UserID          int64
	IsMedia         bool
}
