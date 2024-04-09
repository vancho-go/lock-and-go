package handlers

import (
	"encoding/json"
	"errors"
	api "github.com/vancho-go/lock-and-go/api/http"
	"github.com/vancho-go/lock-and-go/internal/config"
	"github.com/vancho-go/lock-and-go/internal/service/auth"
	"github.com/vancho-go/lock-and-go/internal/service/jwt"
	"github.com/vancho-go/lock-and-go/pkg/customerrors"
	"github.com/vancho-go/lock-and-go/pkg/logger"
	"net/http"
	"time"
)

// UserAuthController контроллер для пользовательской аутентификации.
type UserAuthController struct {
	userService *auth.UserAuthService
	log         *logger.Logger
}

// NewUserController конструктор для UserAuthController.
func NewUserController(userService *auth.UserAuthService, log *logger.Logger) *UserAuthController {
	return &UserAuthController{
		userService: userService,
		log:         log}
}

// Register обработчик регистрации нового пользователя.
func (c *UserAuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req api.RegisterUserRequest
	if err := decodeJSONRequestBody(w, r, &req); err != nil {
		return
	}

	if err := c.userService.Register(r.Context(), req.Username, req.Password); err != nil {
		if errors.Is(err, customerrors.ErrUsernameNotUnique) {
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		}
		c.log.Errorf("failed to register user: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// Authenticate обработчик аутентификации пользователя.
func (c *UserAuthController) Authenticate(w http.ResponseWriter, r *http.Request) {
	var req api.AuthenticateUserRequest
	if err := decodeJSONRequestBody(w, r, &req); err != nil {
		return
	}

	token, err := c.userService.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.As(err, &customerrors.ErrWrongPassword) {
			http.Error(w, "Wrong password", http.StatusUnauthorized)
			return
		}
		c.log.Errorf("error authenticating user: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    jwt.CookieKey,
		Value:   token,
		Expires: time.Now().Add(config.GetJWTTokenDuration()),
		// Важно для безопасности, предотвращает доступ JavaScript к куки
		HttpOnly: true,
		// Куки будет доступна на всех маршрутах
		Path: "/",
		// Куки должна отправляться только по HTTPS
		Secure: true,
		// Предотвращает отправку куки при кросс-доменных запросах
		SameSite: http.SameSiteStrictMode,
	})
}

// decodeJSONRequestBody декодирует тело запроса JSON в предоставленную структуру и проверяет наличие обязательных полей.
func decodeJSONRequestBody(w http.ResponseWriter, r *http.Request, dst interface{}, requiredFields ...string) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return err
	}
	return nil
}
