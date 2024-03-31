package customerrors

import "errors"

var (
	ErrUsernameNotUnique = errors.New("username is already in use")
	ErrWrongPassword     = errors.New("wrong password")
)
