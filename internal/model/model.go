package model

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
