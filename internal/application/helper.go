package application

import (
	"context"
	"github.com/jmoiron/sqlx"
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
