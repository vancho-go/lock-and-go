package jwt

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vancho-go/lock-and-go/internal/config"
	"github.com/vancho-go/lock-and-go/internal/model"
	"net/http"
	"time"
)

// Manager тип, необходимый для работы с JWT секретами.
type Manager struct {
	secretKey     string
	tokenDuration time.Duration
}

// NewJWTManager конструктор Manager.
func NewJWTManager(secretKey string, tokenDuration time.Duration) *Manager {
	return &Manager{secretKey: secretKey, tokenDuration: tokenDuration}
}

// GenerateToken генерирует токен JWT.
func (m *Manager) GenerateToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, model.JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
		},
	})

	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	return tokenString, nil
}

// GetTokenFromCookie извлекает токен из куки запроса.
func GetTokenFromCookie(r *http.Request) (string, error) {
	token, err := r.Cookie(CookieKey)
	if err != nil {
		return "", fmt.Errorf("cookie not found : %w", err)
	}
	if token == nil {
		return "", fmt.Errorf("cookie is empty : %w", err)
	}
	return token.Value, nil
}

// IsTokenValid проверяет токен на валидность.
func IsTokenValid(token string) error {
	verifiedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(config.GetJWTSecretKey()), nil
	})
	if err != nil {
		return err
	}
	if !verifiedToken.Valid {
		return fmt.Errorf("token is not valid")
	}
	return nil
}

// GetUserIDFromToken извлекает UserID из токена.
func GetUserIDFromToken(token string) (string, error) {
	var claims model.JWTClaims
	_, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.GetJWTSecretKey()), nil
	})
	if err != nil {
		return "", fmt.Errorf("error parsing token: %w", err)
	}
	return claims.UserID, nil
}

// GetUserIDFromContext извлекает UserID из контекста.
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(ContextKey).(string)
	return userID, ok
}
