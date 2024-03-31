package api

type RegisterUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthenticateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
