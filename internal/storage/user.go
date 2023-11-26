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
	// Подготовка запроса для вставки данных
	query := `INSERT INTO users (user_id, username) VALUES ($1, $2)`

	// Использование контекста с запросом
	_, err := db.ExecContext(ctx, query, user.UserId, user.Username)
	return err
}

// Функция для добавления кнопки к клавиатуре
func AddKeyboardButton(keyboard tgbotapi.ReplyKeyboardMarkup, newButton string) tgbotapi.ReplyKeyboardMarkup {
	// Создаем новую кнопку
	button := tgbotapi.NewKeyboardButton(newButton)
	maxButtonsPerRow := 3
	// Проверяем, есть ли ряды в клавиатуре
	if len(keyboard.Keyboard) == 0 {
		// Если рядов нет, добавляем новый ряд с кнопкой
		keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(button))
	} else {
		// Если ряды есть, проверяем последний ряд
		lastRowIndex := len(keyboard.Keyboard) - 1
		if len(keyboard.Keyboard[lastRowIndex]) < maxButtonsPerRow {
			// Если в последнем ряду меньше maxButtonsPerRow кнопок, добавляем кнопку в этот ряд
			keyboard.Keyboard[lastRowIndex] = append(keyboard.Keyboard[lastRowIndex], button)
		} else {
			// Если в последнем ряду уже maxButtonsPerRow кнопок, создаем новый ряд
			keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(button))
		}
	}

	return keyboard
}

func AddCategories(ctx context.Context, db *sqlx.DB, userID int64, categoryId int) error {
	//err := db.GetContext(ctx, &userID, "SELECT id FROM users WHERE user_id = $1", userID)
	//if err != nil {
	//	log.Printf("User hasn't registered yet, [ERROR]: %s", err)
	//	return err
	//}
	var categoryID = categoryId
	err := db.GetContext(ctx, &categoryID, "SELECT id FROM categories WHERE category_name = $1", categoryID)

	// Добавляем связь между пользователем и категорией в 'user_categories'
	_, err = db.ExecContext(ctx, "INSERT INTO user_categories (user_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, categoryID)
	return err
}

func RemoveCategories(ctx context.Context, db *sqlx.DB, userID int64, categoryId int) error {
	//err := db.GetContext(ctx, &userID, "SELECT id FROM users WHERE user_id = $1", userID)
	//if err != nil {
	//	log.Printf("User hasn't registered yet, [ERROR]: %s", err)
	//	return err
	//}
	var categoryID = categoryId
	err := db.GetContext(ctx, &categoryID, "SELECT id FROM categories WHERE category_name = $1", categoryID)

	// Добавляем связь между пользователем и категорией в 'user_categories'
	_, err = db.ExecContext(ctx, "DELETE FROM user_categories WHERE user_id= $1 AND category_id= $2", userID, categoryID)
	return err
}

func GetCategoriesMap(ctx context.Context, db *sqlx.DB) (map[int]string, error) {
	categories := make(map[int]string)

	// Запрос для получения всех категорий
	query := `SELECT id, category_name FROM categories`

	var tempCategories []struct {
		ID   int    `db:"id"`
		Name string `db:"category_name"`
	}
	// Выполнение запроса и заполнение временного слайса
	err := db.SelectContext(ctx, &tempCategories, query)
	if err != nil {
		return nil, err // В случае ошибки возвращаем пустую карту и ошибку
	}

	// Заполнение конечной карты данными
	for _, c := range tempCategories {
		categories[c.ID] = c.Name
	}

	return categories, nil // Возвращаем заполненную карту и nil в качестве ошибки
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
