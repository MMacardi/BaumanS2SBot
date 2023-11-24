package storage

import (
	"TelegramS2SBot/internal/model"
	"context"
	"github.com/jmoiron/sqlx"
)

type UserStorage struct {
	db *sqlx.DB
}

func RegisterUser(ctx context.Context, db *sqlx.DB, user model.User) error {
	// Подготовка запроса для вставки данных
	query := `INSERT INTO users (userid, name, category, ranking) VALUES ($1, $2, $3, $4)`

	// Использование контекста с запросом
	_, err := db.ExecContext(ctx, query, user.UserId, user.Name, user.Category, user.Ranking)
	return err
}
