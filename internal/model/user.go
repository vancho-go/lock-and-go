package model

// User модель пользователя.
type User struct {
	Username string
	Password string
}

// UserHashed модель пользователя с хэшированным паролем.
type UserHashed struct {
	ID           string
	Username     string
	PasswordHash string
}
