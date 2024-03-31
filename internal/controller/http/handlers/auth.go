package handlers

import (
	"encoding/json"
	"errors"
	api "github.com/vancho-go/lock-and-go/api/http"
	"github.com/vancho-go/lock-and-go/internal/service/auth"
	"github.com/vancho-go/lock-and-go/pkg/customerrors"
	"github.com/vancho-go/lock-and-go/pkg/logger"
	"net/http"
	"time"
)

type UserAuthController struct {
	userService *auth.UserService
	log         *logger.Logger
}

func NewUserController(userService *auth.UserService, log *logger.Logger) *UserAuthController {
	return &UserAuthController{
		userService: userService,
		log:         log}
}

func (c *UserAuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req api.RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Missing required fields: username, password", http.StatusBadRequest)
		return
	}
	if err := c.userService.Register(r.Context(), req.Username, req.Password); err != nil {
		if errors.As(err, &customerrors.ErrUsernameNotUnique) {
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		}
		c.log.Errorf("register: failed to decode json: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (c *UserAuthController) Authenticate(w http.ResponseWriter, r *http.Request) {
	var req api.AuthenticateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Missing required fields: username, password", http.StatusBadRequest)
		return
	}

	token, err := c.userService.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.As(err, &customerrors.ErrWrongPassword) {
			http.Error(w, "Wrong password", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour), // Установите соответствующий срок действия
		HttpOnly: true,                           // Важно для безопасности, предотвращает доступ JavaScript к куки
		Path:     "/",                            // Куки будет доступна на всех маршрутах
		Secure:   true,                           // Куки должна отправляться только по HTTPS
		SameSite: http.SameSiteStrictMode,        // Предотвращает отправку куки при кросс-доменных запросах
	})
}
