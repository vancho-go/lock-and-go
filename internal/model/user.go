package model

type User struct {
	Username string
	Password string
}

type UserHashed struct {
	ID           string
	Username     string
	PasswordHash string
}
