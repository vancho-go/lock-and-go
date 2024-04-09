package customerrors

import "errors"

var (
	// ErrUsernameNotUnique ошибка при создании пользователя,
	// чей username уже занят.
	ErrUsernameNotUnique = errors.New("username is already in use")
	// ErrWrongPassword ошибка при вводе неверного пароля пользователя.
	ErrWrongPassword = errors.New("wrong password")
)
