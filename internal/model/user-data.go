package model

import "time"

// UserData модель пользовательских данных.
type UserData struct {
	DataID     string    `db:"data_id" json:"data_id"`
	UserID     string    `db:"user_id" json:"user_id"`
	Data       string    `db:"data" json:"data"`
	DataType   string    `db:"data_type" json:"data_type"`
	Status     string    `db:"status" json:"status"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	ModifiedAt time.Time `db:"modified_at" json:"modified_at"`
}
