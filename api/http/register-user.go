package api

// RegisterUserRequest модель запроса на регистрацию пользователя.
type RegisterUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthenticateUserRequest модель запроса на аутентификацию пользователя.
type AuthenticateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
