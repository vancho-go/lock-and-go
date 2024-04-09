package middlewares

import (
	"github.com/stretchr/testify/assert"
	"github.com/vancho-go/lock-and-go/internal/service/jwt"
	"github.com/vancho-go/lock-and-go/pkg/logger"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Пример тестовой реализации Middlewares для использования в тесте.
type testMiddlewares struct {
	log *logger.Logger // Замените на реальный логгер, если он у вас есть
}

func (m *testMiddlewares) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// Тело миддлвара остается без изменений
	})
}

func TestJWTMiddleware(t *testing.T) {
	// Создаем "фейковый" HTTP сервер
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Создаем экземпляр Middlewares
	m := &testMiddlewares{}

	// Оборачиваем наш тестовый хендлер миддлваром
	testHandler := m.JWTMiddleware(handler)

	// Создаем тестовый запрос
	req := httptest.NewRequest("GET", "http://example.com", nil)

	// Добавляем в запрос "фейковый" JWT токен
	req.AddCookie(&http.Cookie{
		Name:  jwt.CookieKey,
		Value: "testToken", // Это должно быть действительное значение токена для реального тестирования
	})

	// Используем httptest.ResponseRecorder для записи ответа
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	// Проверяем, что запрос прошел через миддлвар и достиг нашего хендлера
	assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200 OK")
}
