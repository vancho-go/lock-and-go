package jwt

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestNewJWTManager(t *testing.T) {
	secretKey := "secret"
	tokenDuration := 1 * time.Hour

	manager := NewJWTManager(secretKey, tokenDuration)
	assert.NotNil(t, manager)
	assert.Equal(t, secretKey, manager.secretKey)
	assert.Equal(t, tokenDuration, manager.tokenDuration)
}

func TestGetTokenFromCookie(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	cookieValue := "testToken"
	req.AddCookie(&http.Cookie{Name: CookieKey, Value: cookieValue})

	token, err := GetTokenFromCookie(req)
	assert.NoError(t, err)
	assert.Equal(t, cookieValue, token)
}

func TestGetUserIDFromContext(t *testing.T) {
	userID := "testUser"
	ctx := context.WithValue(context.Background(), ContextKey, userID)

	extractedUserID, ok := GetUserIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, userID, extractedUserID)
}

func TestManager_GenerateToken(t *testing.T) {
	manager := NewJWTManager("secret", time.Minute*5)
	userID := "testUser"

	token, err := manager.GenerateToken(userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}
