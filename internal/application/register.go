package application

import (
	"BaumanS2SBot/internal/application/commands"
	"BaumanS2SBot/internal/application/states"
	"BaumanS2SBot/internal/model"
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"log"
	"strings"
	"time"
)

type UserStorage struct {
	db *sqlx.DB
}

func User(update tgbotapi.Update, ctx context.Context, db *sqlx.DB,
	bot *tgbotapi.BotAPI, userID int64, userStates map[int64]int) {
	if update.Message.Text == "Начать" {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		user := model.User{
			UserId:   userID,
			Username: update.Message.From.UserName,
		}
		if err := AddUserToDB(ctx, db, user); err != nil {
			log.Printf("Error registering user: %v", err)
			cancel()
			return
		}
		cancel()
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добро пожаловать!")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending greeting msg %v", err)
		}
		commands.SendHelpMessage(bot, update.Message.Chat.ID)
		SendHomeKeyboard(bot, update.Message.Chat.ID, userStates, userID, states.StateHome)

	}
}

func AddUserToDB(ctx context.Context, db *sqlx.DB, user model.User) error {
	query := `INSERT INTO users (user_id, username) VALUES ($1, $2)`

	_, err := db.ExecContext(ctx, query, user.UserId, user.Username)
	return err
}

func AddCategories(ctx context.Context, db *sqlx.DB, userID int64, categoryId int) error {
	var categoryID = categoryId
	err := db.GetContext(ctx, &categoryID, "SELECT id FROM categories WHERE category_name = $1", categoryID)

	_, err = db.ExecContext(ctx, "INSERT INTO user_categories (user_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, categoryID)
	return err
}

func RemoveCategories(ctx context.Context, db *sqlx.DB, userID int64, categoryId int) error {
	var categoryID = categoryId
	err := db.GetContext(ctx, &categoryID, "SELECT id FROM categories WHERE category_name = $1", categoryID)

	_, err = db.ExecContext(ctx, "DELETE FROM user_categories WHERE user_id= $1 AND category_id= $2", userID, categoryID)
	return err
}

func GetCategoriesMap(ctx context.Context, db *sqlx.DB) (map[int]string, error) {
	categories := make(map[int]string)

	query := `SELECT id, category_name FROM categories`

	var tempCategories []struct {
		ID   int    `db:"id"`
		Name string `db:"category_name"`
	}

	err := db.SelectContext(ctx, &tempCategories, query)
	if err != nil {
		return nil, err
	}

	for _, c := range tempCategories {
		categories[c.ID] = c.Name
	}
	return categories, nil
}

func GetEveryUserCategoriesMap(ctx context.Context, db *sqlx.DB) (map[int64][]int, error) {
	userCategories := make(map[int64][]int)

	query := `SELECT user_id, category_id FROM user_categories`

	var tempUserCategories []struct {
		UserID     int64 `db:"user_id"`
		CategoryID int   `db:"category_id"`
	}

	err := db.SelectContext(ctx, &tempUserCategories, query)
	if err != nil {
		return nil, err
	}

	for _, c := range tempUserCategories {
		userCategories[c.UserID] = append(userCategories[c.UserID], c.CategoryID)
	}

	return userCategories, nil
}

func GetUserCurrentCategoriesSlice(ctx context.Context, db *sqlx.DB, userID int64) []int {
	EveryUserCategoriesMap, err := GetEveryUserCategoriesMap(ctx, db)
	if err != nil {
		log.Printf("Error getting user categories map: %v", err)
		return nil
	}

	return EveryUserCategoriesMap[userID]
}

func GetCategoriesNameByCategoryID(ctx context.Context, db *sqlx.DB, categoryIDs []int) []string {
	var userCategoryNameSlice []string
	categories, err := GetCategoriesMap(ctx, db)
	if err != nil {
		log.Printf("Error getting categories map: %v", err)
		return nil
	}

	for _, id := range categoryIDs {
		if name, ok := categories[id]; ok {
			userCategoryNameSlice = append(userCategoryNameSlice, name)
		}
	}

	return userCategoryNameSlice
}

func GetCurrentUserCategories(ctx context.Context, db *sqlx.DB, chatID int64) []string {
	currentUserCategories := GetCategoriesNameByCategoryID(ctx, db, GetUserCurrentCategoriesSlice(ctx, db, chatID))

	return currentUserCategories
}

func GetCurrentUserCategoriesString(ctx context.Context, db *sqlx.DB, chatID int64) string {
	userCategoriesString := strings.Join(GetCurrentUserCategories(ctx, db, chatID), ",")
	return userCategoriesString
}

func IsCategoryAdded(ctx context.Context, db *sqlx.DB, chatID int64, input string) bool {
	currentUserCategories := GetCurrentUserCategories(ctx, db, chatID)
	for i := range currentUserCategories {
		if currentUserCategories[i] == input {
			return true
		}
	}
	return false
}

func GetCategories(ctx context.Context, db *sqlx.DB) map[int]string {
	categories, err := GetCategoriesMap(ctx, db)
	if err != nil {
		log.Fatalf("can't take categories map %v", err)
	}
	return categories
}

func GetKeyByValue(myMap map[int]string, valueToFind string) (int, bool) {
	for key, value := range myMap {
		if value == valueToFind {
			return key, true
		}
	}
	return 0, false
}
