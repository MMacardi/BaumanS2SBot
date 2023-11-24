package model

type User struct {
	UserId   int64  `db:"userID"`
	Name     string `db:"name"`
	Category string `db:"category"`
	Ranking  int    `db:"ranking"`
}
