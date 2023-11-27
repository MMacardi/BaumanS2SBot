package storage

import (
	"BaumanS2SBot/internal/model"
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"log"
)

type UserStorage struct {
	db *sqlx.DB
}

func RegisterUser(ctx context.Context, db *sqlx.DB, user model.User) error {
	query := `INSERT INTO users (user_id, username) VALUES ($1, $2)`

	_, err := db.ExecContext(ctx, query, user.UserId, user.Username)
	return err
}

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
		return nil, err // В случае ошибки возвращаем пустую карту и ошибку
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

func GetUserCategoriesSlice(ctx context.Context, db *sqlx.DB, userID int64) []int {
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
